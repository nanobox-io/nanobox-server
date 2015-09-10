package boxfile

import "testing"

// import "fmt"

func Testparse(t *testing.T) {
	box := New([]byte("a: Easy!\nb:\n  c: 2\n  d: [3, 4]\n"))
	if box.Node("a") != "Easy!" {
		t.Error("boxfile parsed does not match boxfile in")
	}

	bad := New([]byte("baz:\n   cdr\nfoo::*)-> bar"))
	if bad.Valid {
		t.Error("Boxfile thinks its valid but it shoudnt be")
	}

}

func TestNode(t *testing.T) {
	box := New([]byte("web1:\n  name: site\n  type: php\n  version: 5.4\n  php_extensions:\n    - mysql\n    - gd\n    - eaccelerator\n"))
	web1 := box.Node("web1")
	if web1.Node("name").(string) != "site" {
		t.Error("nested nodes dont work")
	}
	if string(web1.raw) != "name: site\nphp_extensions:\n- mysql\n- gd\n- eaccelerator\ntype: php\nversion: 5.4\n" {
		t.Error("subnodes dont create raw yaml correctly")
	}
}
