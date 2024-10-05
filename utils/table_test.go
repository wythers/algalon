package utils_test

import (
	"encoding/json"
	"testing"

	"github.com/wythers/algalon/utils"
	"gopkg.in/go-playground/assert.v1"
)

func Test_table(t *testing.T) {
	j := `{
		"inflow":[
			{
				"to": "a46123aaeqeqews",
				"amount": "5",
				"type": "trx",
				"timestamp": 182323141,
				"owner": "a4369e5b-4a11-407f-8f91-a8219fcd9f1a"
			}
		],
		"outflow":[
			{
				"from": "bgdgsgsdgeew",
				"to": "a46123aaeqeqews",
				"amount": "5",
				"type": "trx",
				"txID": "eqeofidsfisforerrqqq",
				"timestamp": 182323141,
				"owner": "a4369e5b-4a11-407f-8f91-a8219fcd9f1a"
			}
		]
	}`

	var record utils.Records
	err := json.Unmarshal([]byte(j), &record)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, len(record.Inflow), 1)
	assert.Equal(t, true,
		record.Inflow[0].To == "a46123aaeqeqews" &&
			record.Inflow[0].Amount == "5" &&
			record.Inflow[0].Type == "trx" &&
			record.Inflow[0].Timestamp == 182323141 &&
			record.Inflow[0].Owner == "a4369e5b-4a11-407f-8f91-a8219fcd9f1a")

	b, err := json.Marshal(record)
	if err != nil {
		t.Error(err)
	}

	ret := `{"inflow":[{"to":"a46123aaeqeqews","amount":"5","type":"trx","timestamp":182323141,"owner":"a4369e5b-4a11-407f-8f91-a8219fcd9f1a"}],"outflow":[{"from":"bgdgsgsdgeew","to":"a46123aaeqeqews","amount":"5","type":"trx","txID":"eqeofidsfisforerrqqq","timestamp":182323141,"owner":"a4369e5b-4a11-407f-8f91-a8219fcd9f1a"}]}`
	assert.Equal(t, ret, string(b))

	var tb utils.Table
	s := tb.Swap(record)
	b, err = json.Marshal(s)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, `{}`, string(b))

	s2 := tb.Swap(utils.Records{})
	b, err = json.Marshal(s2)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, ret, string(b))
}
