package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var nodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "List nodes in inventory-sample.yml",
	RunE: func(cmd *cobra.Command, args []string) error {
		role, _ := cmd.Flags().GetString("role")
		inv, err := loadInventory()
		if err != nil {
			return err
		}
		if role != "" {
			return printNodesByRole(inv, role)
		}
		return printAllNodes(inv)
	},
}

var nodeAddCmd = &cobra.Command{
	Use:   "node-add --role <role> --ip <ip> [flags]",
	Short: "Add one or more nodes to inventory-sample.yml",
	RunE: func(cmd *cobra.Command, args []string) error {
		role, _ := cmd.Flags().GetString("role")
		ips, _ := cmd.Flags().GetStringArray("ip")
		user, _ := cmd.Flags().GetString("user")
		labels, _ := cmd.Flags().GetStringArray("label")

		if role == "" {
			return fmt.Errorf("--role is required (server or agent)")
		}
		if len(ips) == 0 {
			return fmt.Errorf("at least one --ip is required")
		}

		inv, err := loadInventory()
		if err != nil {
			return err
		}

		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if ip == "" {
				continue
			}
			inv.addHost(role, ip, user, labels)
			fmt.Printf("Added %s node: %s\n", role, ip)
		}

		if err := inv.save(); err != nil {
			return err
		}
		fmt.Printf("Saved to %s\n", inventoryFile)
		return nil
	},
}

var nodeRmCmd = &cobra.Command{
	Use:   "node-rm [--ip <address>] [--role <role>] [--all]",
	Short: "Remove nodes from inventory-sample.yml",
	RunE: func(cmd *cobra.Command, args []string) error {
		ip, _ := cmd.Flags().GetString("ip")
		role, _ := cmd.Flags().GetString("role")
		all, _ := cmd.Flags().GetBool("all")

		if ip == "" && role == "" && !all {
			return fmt.Errorf("one of --ip, --role, or --all is required")
		}

		inv, err := loadInventory()
		if err != nil {
			return err
		}

		switch {
		case ip != "":
			if !inv.removeHost(ip) {
				return fmt.Errorf("node '%s' not found in inventory", ip)
			}
			fmt.Printf("Removed node: %s\n", ip)
		case role != "":
			if !inv.removeRole(role) {
				return fmt.Errorf("role '%s' not found in inventory", role)
			}
			fmt.Printf("Removed all %s nodes\n", role)
		case all:
			inv.removeAll()
			fmt.Println("Removed all nodes")
		}

		return inv.save()
	},
}

func printAllNodes(inv *invDoc) error {
	roles := inv.listRoles()
	if len(roles) == 0 {
		fmt.Println("No nodes in inventory.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "IP\tROLE\tUSER")
	fmt.Fprintln(w, "--\t----\t----")
	for _, role := range roles {
		hosts := inv.hostsForRole(role)
		for _, ip := range yKeys(hosts) {
			user := ""
			if h := yGet(hosts, ip); h != nil {
				if u := yGet(h, "ansible_user"); u != nil {
					user = u.Value
				}
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", ip, role, user)
		}
	}
	return w.Flush()
}

func printNodesByRole(inv *invDoc, role string) error {
	hosts := inv.hostsForRole(role)
	ips := yKeys(hosts)
	if len(ips) == 0 {
		fmt.Printf("No nodes for role: %s\n", role)
		return nil
	}

	varName := "extra_" + role + "_args"
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "IP\tUSER\tARGS")
	fmt.Fprintln(w, "--\t----\t----")
	for _, ip := range ips {
		user, extra := "", ""
		if h := yGet(hosts, ip); h != nil {
			if u := yGet(h, "ansible_user"); u != nil {
				user = u.Value
			}
			if a := yGet(h, varName); a != nil {
				extra = a.Value
			}
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", ip, user, extra)
	}
	return w.Flush()
}

func init() {
	nodesCmd.Flags().StringP("role", "r", "", "filter by role (e.g. server, agent)")

	nodeAddCmd.Flags().StringP("role", "r", "", "node role: server or agent (required)")
	nodeAddCmd.Flags().StringArrayP("ip", "i", nil, "node IP address, repeatable for multiple nodes")
	nodeAddCmd.Flags().StringP("user", "u", "root", "SSH remote user")
	nodeAddCmd.Flags().StringArrayP("label", "l", nil, "node label key=value, repeatable")

	nodeRmCmd.Flags().String("ip", "", "remove node by IP address")
	nodeRmCmd.Flags().StringP("role", "r", "", "remove all nodes of this role")
	nodeRmCmd.Flags().Bool("all", false, "remove all nodes from every role")
}
