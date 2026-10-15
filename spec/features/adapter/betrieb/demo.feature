# language: de
@spec-only @integration
Funktionalität: Reale Demo und ehrliche Performanceabnahme
  Als Betreiber starte ich echte Dienste und unterscheide HTTPmessung vom sichtbaren Browseraufbau.

  Szenario: Persistente echte Dienste tragen den Gastkauf
    Angenommen die dokumentierte Composekonfiguration startet reale Dienste
      | Service | RequiredCapability |
      | PostgreSQL | transactions,logical-pgoutput,outbox |
      | Temporal | cart-and-order-history |
      | OpenSearch | nested-query-and-external-version |
      | Nginx | SSI |
      | Varnish | ESI |
    Wenn ein Gastkauf abgeschlossen und die Laufzeit neu gestartet wird
      | PaymentMethod | ShippingMethod | Restart | CommandID |
      | invoice | standard | worker-and-services-preserve-volumes | cmd-demo |
    Dann lautet die reale Integrationsbeobachtung
      | Orders | CartWorkflows | OrderWorkflows | ReservationPresent | SnapshotRetained | RealExternalPayment |
      | 1 | 1 | 1 | true | true | false |

  Szenariogrundriss: Messberichte besitzen eine eindeutige Grenze
    Angenommen der dokumentierte Messdatensatz ist geladen
      | Products | Variants | Categories | Profiles | Orders | Concurrency | Network | MinimumCPU | MinimumRAMGiB |
      | 30 | 90 | 5 | 4 | 100 | 10 | loopback | 4 | 8 |
    Wenn die vereinbarten Anzeigerouten gemessen werden
      | Boundary | Cache | WarmupPerRoute | RequestsPerRoute | Routes |
      | <grenze> | <cache> | <warmup> | <anzahl> | home,category,search,product,cart,checkout,order |
    Dann muss der Messbericht die Aussagegrenze enthalten
      | TargetP95Milliseconds | RequiredStatistics | RequiredEnvironment | ProvesBrowserLCP | MutatingCommandsIncluded |
      | 100 | p50,p95,p99,errorRate | CPU,RAM,OS,images,dataset,TCP,compression | <lcp> | false |
    Beispiele:
      | grenze | cache | warmup | anzahl | lcp |
      | requestStart-to-last-composed-HTML-byte | warm | 100 | 1000 | false |
      | requestStart-to-last-composed-HTML-byte | cold-cache-services-ready | 0 | 100 | false |
      | navigationStart-to-visible-LCP | browser-smoke | 0 | 1 | true |

  Szenario: Fehlende Browsermessung bleibt offen statt HTTPersatz
    Wenn nur zusammengesetztes HTTP gemessen wurde
      | HTTPP95Milliseconds | BrowserLCPMeasured |
      | 80 | false |
    Dann lautet die Leistungsbehauptung
      | HTTPGoal | BrowserGoal | CompletePageSLAAchieved |
      | met-in-measured-profile | unbelegt | false |
