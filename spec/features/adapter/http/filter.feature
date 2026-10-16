# language: de
@http
Funktionalität: Typisierte teilbare Filtereingaben
  Als Gast übergebe ich Marken Optionen und exakte Intervalle ohne eigene Suchlogik der Oberfläche.

  Hintergrund:
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | |
    Und HTTPportdaten werden ersetzt
      | JSON |
      | {"Category":{"FilterDefinitions":[{"Code":"storageCapacity","Type":"number","Level":"variant","Name":{"de-DE":"Speicher"}}]}} |

  Szenariogrundriss: Gültige Intervalle werden als typisierte Applicationquery ausgeführt
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /de-DE/kategorie/smartphones--c-phone?<query> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
    Beispiele:
      | query |
      | storageCapacity.min=128&storageCapacity.max=256 |
      | priceMin=10000&priceMax=14000 |
      | priceMax=14000 |
      | priceMin=10000 |
      | priceMin=&priceMax=&page=1&sort=relevance&view=grid |
      | sort=nameAsc&view=list |

  Szenariogrundriss: Ungültige oder mehrdeutige Eingaben bleiben sichtbar ungültig
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /de-DE/kategorie/smartphones--c-phone?<query> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 422 |
    Beispiele:
      | query |
      | storageCapacity.min=invalid |
      | storageCapacity.min=128&storageCapacity.min=256 |
      | storageCapacity.min=256&storageCapacity.max=128 |
      | priceMin=invalid |
      | priceMax=invalid |
      | priceMin=-1 |
      | priceMin=14000&priceMax=10000 |
      | a=1&b=1&c=1&d=1&e=1 |
      | q=first&q=second |
      | q=%ZZ |

  Szenario: Ungültiger Filter wird auch vor einem Slugredirect abgewiesen
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /de-DE/kategorie/alt--c-phone?redirect=https://foreign.example |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 422 |
      | AbsentHeader:Location | https://foreign.example |

  Szenario: Die Listenansicht und Intervalle bleiben sichtbar ausgewählt
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /de-DE/kategorie/smartphones--c-phone?storageCapacity.min=128&storageCapacity.max=256&priceMin=10000&priceMax=14000&view=list |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | BodyAll | class="list",value="list" selected,name="storageCapacity.min",value="128",name="storageCapacity.max",value="256",name="priceMin",value="10000",name="priceMax",value="14000" |
