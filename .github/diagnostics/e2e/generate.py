#!/usr/bin/env python3
"""Einmaliger, an den geprüften Quellstand gebundener Go-Overlayaufbau."""
import hashlib
import json
from pathlib import Path
import subprocess
import sys

root = Path.cwd()
output = Path(sys.argv[1]).resolve()
output.mkdir(parents=True, exist_ok=True, mode=0o700)
files = {}

def source(path):
    if path not in files:
        files[path] = (root / path).read_text()
    return files[path]

def change(path, before, after):
    text = source(path)
    if text.count(before) != 1:
        raise SystemExit("overlay anchor mismatch: " + path)
    files[path] = text.replace(before, after)

def imported(path, extra=()):
    text = source(path)
    additions = ['"github.com/Alps86/shop/internal/e2ediag"'] + list(extra)
    files[path] = text.replace("import (", "import (\n" + "\n".join(additions), 1)

def wrap(path, signature, profile):
    names, region, call = profile
    original, replacement = names.split("|")
    change(path, signature, signature.replace(") "+original+"(", ") "+replacement+"(", 1))
    prefix = "ctx, finish := e2ediag.Begin(request.Context(), "+json.dumps(region)+")\n"
    files[path] += "\n"+signature+"\n"+prefix+"defer finish()\nreturn "+call+"\n}\n"

main = "cmd/shop/main.go"
imported(main)
change(main, "func main() {", 'func main() {\nfinishDiagnostic, diagnosticError := e2ediag.Bootstrap()\nif diagnosticError != nil {\n log.Print("diagnostic-bootstrap-unavailable")\n os.Exit(1)\n}\ndefer finishDiagnostic()')

handler = "internal/adapter/http/handler.go"
imported(handler, ('"context"', '"io"'))
change(handler, 'func (handler *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {', 'func (handler *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {\n request, finish := e2ediag.Request(request, writer)\n defer finish()')
change(handler, 'func (handler *Handler) render(', 'func (handler *Handler) renderUnprofiled(')
change(handler, 'handler.renderer.templates.ExecuteTemplate(&output, name, page)', 'handler.renderer.profileTemplate(request.Context(), &output, name, page)')
change(handler, 'page.write(writer, output.Bytes())', 'page.profileWrite(request.Context(), writer, output.Bytes())')
files[handler] += '''
func (handler *Handler) render(writer http.ResponseWriter, request *http.Request, page Page) {
 ctx, finish := e2ediag.Begin(request.Context(), "http.prepare-render")
 defer finish()
 handler.renderUnprofiled(writer, request.WithContext(ctx), page)
}

func (renderer Renderer) profileTemplate(ctx context.Context, writer io.Writer, name string, page Page) error {
 _, finish := e2ediag.Begin(ctx, "http.template")
 defer finish()
 return renderer.templates.ExecuteTemplate(writer, name, page)
}

func (page Page) profileWrite(ctx context.Context, writer http.ResponseWriter, body []byte) {
 _, finish := e2ediag.Begin(ctx, "http.response-write")
 defer finish()
 page.write(writer, body)
}
'''
for path, signature, names, region, call in [
 ("internal/adapter/http/query.go", "func (handler *Handler) query(request *http.Request, schema application.SearchSchema) (application.SearchQuery, error) {", "query|queryUnprofiled", "http.query-parse", "handler.queryUnprofiled(request.WithContext(ctx), schema)"),
 ("internal/adapter/http/catalog.go", "func (handler *Handler) categoryQuery(request *http.Request, category application.CategoryView) (application.SearchQuery, error) {", "categoryQuery|categoryQueryUnprofiled", "http.query-and-user", "handler.categoryQueryUnprofiled(request.WithContext(ctx), category)"),
 ("internal/adapter/http/interest.go", "func (handler *Handler) requestInterests(request *http.Request) ([]sale.CategoryInterest, error) {", "requestInterests|requestInterestsUnprofiled", "http.user-interests", "handler.requestInterestsUnprofiled(request.WithContext(ctx))"),
]:
    imported(path)
    wrap(path, signature, (names, region, call))

for path, needle, phase in [
 ("internal/adapter/postgres/guest.go", '"SELECT s.user_id,s.issued_at,u.profile', "sql.resolve-session"),
 ("internal/adapter/postgres/guest.go", '"SELECT id FROM carts', "sql.active-cart"),
 ("internal/adapter/postgres/interests.go", '"SELECT payload FROM users', "sql.user-payload"),
]:
    if 'e2ediag"' not in source(path):
        imported(path)
    line = next(line for line in source(path).splitlines() if needle in line)
    operation = line.strip().removeprefix("err := ")
    replacement = 'err := e2ediag.SQL(ctx, '+json.dumps(phase)+', func(ctx context.Context) error {\n return '+operation+'\n})'
    change(path, line, replacement)

store = "internal/adapter/postgres/store.go"
imported(store)
change(store, "pgxpool.NewWithConfig(ctx, poolConfig)", "e2ediag.OpenPool(ctx, poolConfig)")

adapter = "internal/adapter/opensearch/adapter.go"
imported(adapter)
start = source(adapter).index("func (adapter Adapter) Search(")
end = source(adapter).index("\nfunc (adapter Adapter) performSearch", start)
files[adapter] = source(adapter)[:start]+'''
func (adapter Adapter) Search(ctx context.Context, query application.SearchQuery) (application.SearchResult, error) {
 ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
 defer cancel()
 body, err := adapter.profilePreparation(ctx, query)
 if err != nil {
  return application.SearchResult{}, err
 }
 return adapter.performSearch(ctx, query, body)
}

func (adapter Adapter) profilePreparation(ctx context.Context, query application.SearchQuery) (JSON, error) {
 _, finish := e2ediag.Begin(ctx, "os.query-build-validate")
 defer finish()
 body, err := adapter.searchBody(query)
 if err != nil {
  return nil, err
 }
 return body, query.Validate(adapter.config.Schema)
}
'''+source(adapter)[end:]
change(adapter, "adapter.searchResult(reply, query, request.Multi)", "adapter.searchResult(ctx, reply, query, request.Multi)")
change(adapter, "func (adapter Adapter) searchResult(reply", "func (adapter Adapter) searchResult(ctx context.Context, reply")
change(adapter, "decodeSearchResponse(reply.Body, multi)", "decodeSearchResponse(ctx, reply.Body, multi)")
change(adapter, "return response.searchResult(query, multi)", 'return response.profileResult(ctx, query, multi)')
files[adapter] += '''
func (response SearchResponse) profileResult(ctx context.Context, query application.SearchQuery, multi bool) (application.SearchResult, error) {
 _, finish := e2ediag.Begin(ctx, "os.cards-facets-suggestions")
 defer finish()
 return response.searchResult(query, multi)
}
'''

transport = "internal/adapter/opensearch/transport.go"
imported(transport)
change(transport, "body, err := encodePayload(payload)", '_, finish := e2ediag.Begin(ctx, "os.encode")\nbody, err := encodePayload(payload)\nfinish()')
change(transport, "adapter.config.HTTP.Do(request)", "e2ediag.HTTP(ctx, adapter.config.HTTP, request)")
change(transport, "adapter.receive(response)", "adapter.receive(ctx, response)")
change(transport, "func (adapter Adapter) receive(response *http.Response) (Reply, error) {", 'func (adapter Adapter) receive(ctx context.Context, response *http.Response) (Reply, error) {\n _, finish := e2ediag.Begin(ctx, "os.http-body")\n defer finish()')

multi = "internal/adapter/opensearch/msearch.go"
imported(multi, ('"context"',))
change(multi, "func decodeSearchResponse(body []byte, multi bool) (SearchResponse, error) {", 'func decodeSearchResponse(ctx context.Context, body []byte, multi bool) (SearchResponse, error) {\n ctx, finish := e2ediag.Begin(ctx, "os.decode-envelope")\n defer finish()')
change(multi, "decodeMultiSearch(body)", "decodeMultiSearch(ctx, body)")
change(multi, "func decodeMultiSearch(body", "func decodeMultiSearch(ctx context.Context, body")
change(multi, "return response.result()", 'result, err := response.result()\nif err == nil {\ne2ediag.Took(ctx, "os.outer-took-ms", response.DiagnosticTook)\ne2ediag.Took(ctx, "os.main-took-ms", response.Responses[0].DiagnosticTook)\ne2ediag.Took(ctx, "os.completion-took-ms", response.Responses[1].DiagnosticTook)\n}\nreturn result, err')
change(multi, "return response, err", 'if err == nil {\n e2ediag.Took(ctx, "os.main-took-ms", response.DiagnosticTook)\n}\nreturn response, err')

types = "internal/adapter/opensearch/type.go"
change(types, "type SearchResponse struct {", 'type SearchResponse struct {\nDiagnosticTook *int64 `json:"took"`')
change(types, "type multiSearchResponse struct {", 'type multiSearchResponse struct {\nDiagnosticTook *int64 `json:"took"`')

for template in (root / ".github/diagnostics/e2e/templates").glob("*.go.tmpl"):
    files["internal/e2ediag/"+template.name.removesuffix(".tmpl")] = template.read_text()

mapping = {}
manifest = []
for relative, content in files.items():
    target = output / relative
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_text(content)
    subprocess.run(["gofmt", "-w", str(target)], check=True)
    mapping[str(root / relative)] = str(target)
    manifest.append({"source":relative, "sha256":hashlib.sha256(target.read_bytes()).hexdigest()})
(output / "overlay.json").write_text(json.dumps({"Replace":mapping}, indent=2))
(output / "manifest.json").write_text(json.dumps(manifest, indent=2))
