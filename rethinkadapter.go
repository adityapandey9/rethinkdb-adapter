package rethinkadapter

import (
	"fmt"
	"runtime"

	"github.com/casbin/casbin/model"
	"github.com/casbin/casbin/persist"
	r "gopkg.in/gorethink/gorethink.v3"
)

// adapter represents the RethinkDB adapter for policy storage.
type adapter struct {
	session r.QueryExecutor
}

type policy struct {
	ID    string `gorethink:"id,omitempty"`
	PTYPE string `gorethink:"ptype"`
	V1    string `gorethink:"v1"`
	V2    string `gorethink:"v2"`
	V3    string `gorethink:"v3"`
	V4    string `gorethink:"v4"`
}

func finalizer(a *adapter) {
	a.close()
}

// NewAdapter is the constructor for adapter.
func NewAdapter(Sessionvar r.QueryExecutor) persist.Adapter {
	a := &adapter{session: Sessionvar}
	a.open()
	// Call the destructor when the object is released.
	runtime.SetFinalizer(a, finalizer)
	return a
}

func (a *adapter) close() {
	a.session = nil
}

func (a *adapter) createDatabase() error {
	_, err := r.DBList().Contains("casbin").Do(r.DBCreate("casbin").Exec(a.session)).Run(a.session)
	if err != nil {
		return err
	}
	return nil
}

func (a *adapter) createTable() error {
	_, err := r.DB("casbin").TableList().Contains("policy").Do(r.DB("casbin").TableCreate("policy").Exec(a.session)).Run(a.session)
	if err != nil {
		return err
	}
	return nil
}

func (a *adapter) open() {
	if err := a.createDatabase(); err != nil {
		panic(err)
	}

	if err := a.createTable(); err != nil {
		panic(err)
	}
}

//Erase the table data
func (a *adapter) dropTable() error {
	_, err := r.DB("casbin").Table("policy").Delete().Run(a.session)
	if err != nil {
		panic(err)
	}
	return nil
}

func loadPolicyLine(line policy, model model.Model) {
	if line.PTYPE == "" {
		return
	}

	key := line.PTYPE
	sec := key[:1]

	tokens := []string{}

	if line.V1 != "" {
		tokens = append(tokens, line.V1)
	}

	if line.V2 != "" {
		tokens = append(tokens, line.V2)
	}

	if line.V3 != "" {
		tokens = append(tokens, line.V3)
	}

	if line.V4 != "" {
		tokens = append(tokens, line.V4)
	}

	model[sec][key].Policy = append(model[sec][key].Policy, tokens)
}

// LoadPolicy loads policy from database.
func (a *adapter) LoadPolicy(model model.Model) error {
	a.open()

	rows, errn := r.DB("casbin").Table("policy").Run(a.session)
	if errn != nil {
		fmt.Printf("E: %v\n", errn)
		return errn
	}

	defer rows.Close()
	var output policy

	for rows.Next(&output) {
		loadPolicyLine(output, model)
	}
	return nil
}

func (a *adapter) writeTableLine(ptype string, rule []string) policy {
	items := policy{
		PTYPE: ptype,
	} //map[string]string{"PTYPE": ptype, "V1": "", "V2": "", "V3": "", "V4": ""}
	for i := 0; i < len(rule); i++ {
		switch i {
		case 0:
			items.V1 = rule[i]
		case 1:
			items.V2 = rule[i]
		case 2:
			items.V3 = rule[i]
		case 3:
			items.V4 = rule[i]
		}
	}
	return items
}

// SavePolicy saves policy to database.
func (a *adapter) SavePolicy(model model.Model) error {
	a.open()
	a.dropTable()
	var lines []policy

	for PTYPE, ast := range model["p"] {
		for _, rule := range ast.Policy {
			line := a.writeTableLine(PTYPE, rule)
			lines = append(lines, line)
		}
	}

	for PTYPE, ast := range model["g"] {
		for _, rule := range ast.Policy {
			line := a.writeTableLine(PTYPE, rule)
			lines = append(lines, line)
		}
	}
	_, err := r.DB("casbin").Table("policy").Insert(lines).Run(a.session)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return err
	}
	return nil
}

//AddPolicy for adding a new policy to rethinkdb
func (a *adapter) AddPolicy(sec string, PTYPE string, policys []string) error {
	line := a.writeTableLine(PTYPE, policys)
	_, err := r.DB("casbin").Table("policy").Insert(line).Run(a.session)
	if err != nil {
		return err
	}
	return nil
}

//RemovePolicy for removing a policy rule from rethinkdb
func (a *adapter) RemovePolicy(sec string, PTYPE string, policys []string) error {
	line := a.writeTableLine(PTYPE, policys)
	_, err := r.DB("casbin").Table("policy").Filter(line).Delete().Run(a.session)
	if err != nil {
		return err
	}
	return nil
}

//RemoveFilteredPolicy for removing filtered policy
func (a *adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	var selector policy
	selector.PTYPE = ptype

	if fieldIndex <= 0 && 0 < fieldIndex+len(fieldValues) {
		selector.V1 = fieldValues[0-fieldIndex]
	}
	if fieldIndex <= 1 && 1 < fieldIndex+len(fieldValues) {
		selector.V2 = fieldValues[1-fieldIndex]
	}
	if fieldIndex <= 2 && 2 < fieldIndex+len(fieldValues) {
		selector.V3 = fieldValues[2-fieldIndex]
	}
	if fieldIndex <= 3 && 3 < fieldIndex+len(fieldValues) {
		selector.V4 = fieldValues[3-fieldIndex]
	}

	_, err := r.DB("casbin").Table("policy").Filter(selector).Delete().Run(a.session)
	if err != nil {
		panic(err)
	}
	return nil
}
