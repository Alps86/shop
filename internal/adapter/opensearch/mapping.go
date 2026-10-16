package opensearch

const mapping = `{
 "settings":{"number_of_shards":1,"number_of_replicas":0},
 "mappings":{
  "dynamic":"strict",
  "dynamic_templates":[
   {"option_values":{"path_match":"variants.options.*","mapping":{"type":"keyword"}}},
   {"number_values":{"path_match":"variants.numbers.*","mapping":{"type":"long"}}}
  ],
  "properties":{
   "product_id":{"type":"keyword","store":true},"revision":{"type":"long"},"deleted":{"type":"boolean"},"published":{"type":"boolean"},
   "name":{"type":"text","analyzer":"german","fields":{"sort":{"type":"keyword"}}},
   "description":{"type":"text","analyzer":"german"},"brand":{"type":"keyword","fields":{"text":{"type":"text","analyzer":"german"}}},
   "category_ids":{"type":"keyword"},"suggest":{"type":"completion","contexts":[{"name":"visibility","type":"category","path":"visibility"}]},"visibility":{"type":"keyword"},"words":{"type":"completion","contexts":[{"name":"visibility","type":"category","path":"visibility"}]},
   "variants":{"type":"nested","dynamic":false,"properties":{
    "margin":{"type":"long"},"sku":{"type":"keyword","store":true},"price":{"type":"long"},"available":{"type":"long"},
    "options":{"type":"object","dynamic":true},"numbers":{"type":"object","dynamic":true},"card":{"type":"object","enabled":false},"card_payload":{"type":"binary","store":true,"doc_values":false}
   }}
  }
 }
}`
