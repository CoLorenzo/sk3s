package cmd

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const inventoryFile = "inventory-sample.yml"

// invDoc wraps a parsed YAML document for in-place mutation.
type invDoc struct {
	doc *yaml.Node
}

func loadInventory() (*invDoc, error) {
	data, err := os.ReadFile(inventoryFile)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("reading %s: %w", inventoryFile, err)
	}

	var doc yaml.Node
	if len(data) > 0 {
		if err := yaml.Unmarshal(data, &doc); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", inventoryFile, err)
		}
	}
	if doc.Kind == 0 {
		doc = yaml.Node{
			Kind: yaml.DocumentNode,
			Content: []*yaml.Node{
				{Kind: yaml.MappingNode, Tag: "!!map"},
			},
		}
	}
	return &invDoc{doc: &doc}, nil
}

func (d *invDoc) save() error {
	out, err := yaml.Marshal(d.doc)
	if err != nil {
		return err
	}
	return os.WriteFile(inventoryFile, out, 0644)
}

func (d *invDoc) root() *yaml.Node {
	if d.doc.Kind == yaml.DocumentNode && len(d.doc.Content) > 0 {
		return d.doc.Content[0]
	}
	return d.doc
}

func (d *invDoc) children() *yaml.Node {
	k3s := yGet(d.root(), "k3s_cluster")
	if k3s == nil {
		return nil
	}
	return yGet(k3s, "children")
}

func (d *invDoc) ensureChildren() *yaml.Node {
	k3s := yEnsure(d.root(), "k3s_cluster")
	return yEnsure(k3s, "children")
}

func (d *invDoc) listRoles() []string {
	return yKeys(d.children())
}

func (d *invDoc) hostsForRole(role string) *yaml.Node {
	ch := d.children()
	if ch == nil {
		return nil
	}
	roleNode := yGet(ch, role)
	if roleNode == nil {
		return nil
	}
	return yGet(roleNode, "hosts")
}

func (d *invDoc) addHost(role, ip, user string, labels []string) {
	ch := d.ensureChildren()
	roleNode := yEnsure(ch, role)
	hosts := yEnsure(roleNode, "hosts")
	hostNode := yEnsure(hosts, ip)

	if user != "" {
		ySet(hostNode, "ansible_user", yScalar(user))
	}
	varName := "extra_" + role + "_args"
	ySet(hostNode, varName, yScalar(buildNodeArgs(labels)))
}

func (d *invDoc) removeHost(ip string) bool {
	ch := d.children()
	if ch == nil {
		return false
	}
	found := false
	for _, role := range yKeys(ch) {
		roleNode := yGet(ch, role)
		if roleNode == nil {
			continue
		}
		hosts := yGet(roleNode, "hosts")
		if hosts == nil {
			continue
		}
		if yGet(hosts, ip) != nil {
			yDel(hosts, ip)
			found = true
		}
	}
	return found
}

func (d *invDoc) removeRole(role string) bool {
	ch := d.children()
	if ch == nil || yGet(ch, role) == nil {
		return false
	}
	yDel(ch, role)
	return true
}

func (d *invDoc) removeAll() {
	k3s := yGet(d.root(), "k3s_cluster")
	if k3s != nil {
		yDel(k3s, "children")
	}
}

func buildNodeArgs(labels []string) string {
	var parts []string
	for _, l := range labels {
		l = strings.TrimSpace(l)
		if l != "" {
			parts = append(parts, "--node-label "+l)
		}
	}
	return strings.Join(parts, " ")
}

// --- yaml.Node low-level helpers ---

func yGet(m *yaml.Node, key string) *yaml.Node {
	if m == nil || m.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}

func ySet(m *yaml.Node, key string, val *yaml.Node) {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			m.Content[i+1] = val
			return
		}
	}
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key},
		val,
	)
}

func yDel(m *yaml.Node, key string) {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			m.Content = append(m.Content[:i], m.Content[i+2:]...)
			return
		}
	}
}

func yKeys(m *yaml.Node) []string {
	if m == nil || m.Kind != yaml.MappingNode {
		return nil
	}
	keys := make([]string, 0, len(m.Content)/2)
	for i := 0; i < len(m.Content); i += 2 {
		keys = append(keys, m.Content[i].Value)
	}
	return keys
}

func yEnsure(m *yaml.Node, key string) *yaml.Node {
	if v := yGet(m, key); v != nil && v.Kind == yaml.MappingNode {
		return v
	}
	newMap := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	ySet(m, key, newMap)
	return newMap
}

func yScalar(value string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Value: value}
}
