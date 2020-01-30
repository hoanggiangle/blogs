package ssql

import (
	"reflect"
	"strings"
	"testing"
)

func TestDbRange(t *testing.T) {
	ranges := []string{
		"1",
		"1,3-5",
		"1,3-5,6",
		"1,",
		",",
		"",
		"AlL",
	}
	for _, r := range ranges {
		_, isAll, err := ParseListDatabaseID(r)
		if err != nil {
			t.Fatal(err)
		}
		if strings.ToLower(string(r)) == "all" && !isAll {
			t.Fatalf("Parse %s must be all", r)
		}
	}
}

func TestDbRangeError(t *testing.T) {
	ranges := []string{
		"must error",
	}
	for _, r := range ranges {
		_, _, err := ParseListDatabaseID(r)
		if err == nil {
			t.Fatal("Must error")
		}
	}
}

func TestSqlManagerMember(t *testing.T) {
	s := `
members:
- servers: [test]
  dbrange: 1-5,9
  dbrange_exhausted: all`

	c, err := loadSqlManagerConfig([]byte(s))
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(c.Members[0].GetDBRange(), c.Members[0].GetExhaustedDBRange()) {
		t.Fatal("DBRange & ExhaustedDBRange must be same when 'all'")
	}
}

func TestSqlManagerMemberSubset(t *testing.T) {
	s := `
members:
- servers: [test]
  dbrange: 1-5,9
  dbrange_exhausted: 3-4`

	_, err := loadSqlManagerConfig([]byte(s))
	if err != nil {
		t.Fatal(err)
	}
}

func TestSqlManagerMemberNotSubset(t *testing.T) {
	s := `
members:
- servers: [test]
  dbrange: 1-5,9
  dbrange_exhausted: 1-2,100`

	_, err := loadSqlManagerConfig([]byte(s))
	if err == nil {
		t.Fatal("Must have error")
	}
	if err != DBRangeErrorNotSubSet {
		t.Fatal(err)
	}
}

func TestSqlManagerMember3(t *testing.T) {
	s := `
members:
- servers: [test]
  dbrange: all`

	_, err := loadSqlManagerConfig([]byte(s))
	if err == nil {
		t.Fatal("Must have error")
	}
	if err != DBRangeErrorAll {
		t.Fatal(err)
	}
}
