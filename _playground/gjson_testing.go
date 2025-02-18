package main

import (
	"fmt"

	"github.com/tidwall/gjson"
)

func main() {
	var msg string = `{
    "data": [
      {
        "ID": "process:Lohnabrechnung",
        "title": "Lohnabrechnung",
        "description": "Verarbeitung der L?hne der MA",
        "created_at": "2025-01-28T09:35:11.308358147Z",
        "updated_at": "2025-01-28T09:35:11.308222574Z",
        "meep": {"foo": "bar"},
        "responsible": [
          {
            "ID": "entity:carsten_jockel",
            "abbrev": "",
            "first_name": "Carsten",
            "last_name": "Jockel",
            "address_street": "An der Kirche",
            "address_number": "2",
            "address_zip_code": 35463,
            "address_city": "Fernwald",
            "address_country": "Deutschland",
            "address_extra": "",
            "phone": "+4964046580351",
            "mobile": "+4917623219765",
            "email": "carsten.jockel@senpro.it",
            "is_external": false,
            "is_company": true,
            "organization": "SenproIT GmbH",
            "position": "Gesch?ftsf?hrer"
          }
        ],
        "legal_basis": [
          {
            "ID": "legal_basis:Arbeitsvertrag",
            "title": "Arbeitsvertrag",
            "kind": "Arbeitsvertrag",
            "document": null,
            "notes": ""
          }
        ],
        "risks": [
          {
            "ID": "risk:Datenverlust",
            "title": "Datenverlust durch Systemfehler",
            "description": "Datenverlust welcher durch fehlerhafte Systeme verursacht werden - z.B. Programmabsurtz, Datenbankfehler.",
            "probably_damage": 0.25,
            "probably_happens": 0.1
          },
          {
            "ID": "risk:CyberDiebstahl",
            "title": "Daten-Diebstahl durch Cyber-Angriff",
            "description": "Verlust der Daten durch Ransomware o.?.",
            "probably_damage": 1,
            "probably_happens": 0.2
          }
        ],
        "storage": [
          {
            "ID": "storage:68g768totaloiyadcfhi",
            "duration": "1 year",
            "reason": {
              "ID": "legal_basis:Arbeitsvertrag",
              "title": "Arbeitsvertrag",
              "kind": "Arbeitsvertrag",
              "document": null,
              "notes": ""
            },
            "location": "Keller"
          }
        ],
        "tasks": [
          {
            "ID": "task:Einsicht_SAP",
            "title": "Einsicht in die gespeicherten Informationen in SAP",
            "description": "Verwendung von SAP um die notwendigen Informationen zu erheben."
          },
          {
            "ID": "task:Bankueberweisung",
            "title": "Einreichung einer Bank?berweisung",
            "description": "Aufbau und Einreichung einer Bank?berweisung per SEPA o.?.; "
          }
        ],
        "affected_entitys": [
          {
            "ID": "entity_kind:Mitarbeiter",
            "description": "Mitarbeiter im Unternehmen"
          }
        ],
        "affected_data": [
          {
            "ID": "data_kind_group:PII",
            "description": "Personen-bezogene Informationen",
            "kinds": [
              {
                "ID": "data_kind:Name",
                "description": "Vor- und Nachname"
              },
              {
                "ID": "data_kind:Adresse",
                "description": "Wohn-/Postanschrift"
              },
              {
                "ID": "data_kind:Geschlecht",
                "description": "Geschlechtsangabe"
              }
            ]
          }
        ]
      }
    ]
  }`
	original := gjson.Parse(msg)
	data := gjson.Parse(original.Get("data").Array()[0].Raw)
	var grabKeys func(gjson.Result) []string
	grabKeys = func(o gjson.Result) []string {
		out := []string{}
		o.ForEach(func(key, value gjson.Result) bool {
			//fmt.Println(key, "-> ", value.Path(original.Raw))
			out = append(out, value.Path(data.Raw))
			if value.Type == gjson.JSON {
				out = append(out, grabKeys(value)...)
			} // we ignore arrays, for now.
			return true
		})
		return out
	}
	fmt.Println(grabKeys(data))
}
