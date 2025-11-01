package node

import (
	"context"
	"errors"
	"fmt"
	"github.com/mertbahardogan/escope/cmd/core"
	"github.com/mertbahardogan/escope/internal/connection"
	"github.com/mertbahardogan/escope/internal/constants"
	"github.com/mertbahardogan/escope/internal/elastic"
	"github.com/mertbahardogan/escope/internal/models"
	"github.com/mertbahardogan/escope/internal/services"
	"github.com/mertbahardogan/escope/internal/ui"
	"github.com/mertbahardogan/escope/internal/util"
	"github.com/spf13/cobra"
	"strings"
)

var nodeCmd = &cobra.Command{
	Use:           "node",
	Short:         "Show node information with health summary and JVM heap details",
	SilenceErrors: true,
	Run: func(cmd *cobra.Command, args []string) {
		client := elastic.NewClientWrapper(connection.GetClient())
		nodeService := services.NewNodeService(client)

		nodes, err := util.ExecuteWithTimeout(func() ([]models.NodeInfo, error) {
			return nodeService.GetNodesInfo(context.Background())
		})
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Printf("Node info check failed: %s\n", constants.MsgTimeoutGeneric)
			} else {
				fmt.Printf("Node info check failed: %v\n", err)
			}
			return
		}

		headers := []string{"Roles", "CPU%", "Mem%", "Heap%", "Disk%", "Free Disk", "Total Disk", "Docs", "Heap Used", "Heap Max", "IP", "Name"}
		rows := make([][]string, 0, len(nodes))

		for _, node := range nodes {
			var filteredRoles []string
			for _, role := range node.Roles {
				if role == "data" || role == "master" {
					filteredRoles = append(filteredRoles, role)
				}
			}

			roles := constants.DashString
			if len(filteredRoles) > 0 {
				roles = strings.Join(filteredRoles, ",")
			}

			name := constants.DashString
			if node.Name != "" {
				name = node.Name
			}

			memPercent := constants.DashString
			if node.MemPercent != "" {
				memPercent = node.MemPercent
			}

			diskPercent := constants.DashString
			if node.DiskPercent != "" {
				diskPercent = node.DiskPercent
			}

			diskTotal := constants.DashString
			if node.DiskTotal != "" {
				diskTotal = node.DiskTotal
			}

			heapUsed := constants.DashString
			if node.HeapUsed != "" {
				heapUsed = node.HeapUsed
			}

			heapMax := constants.DashString
			if node.HeapMax != "" {
				heapMax = node.HeapMax
			}

			docsStr := util.FormatDocsCount(node.Documents)

			if len(roles) > 13 {
				roles = roles[:10] + "..."
			}
			if len(heapUsed) > 10 {
				heapUsed = heapUsed[:7] + "..."
			}
			if len(heapMax) > 10 {
				heapMax = heapMax[:7] + "..."
			}

			row := []string{
				roles,
				node.CPUPercent,
				memPercent,
				node.HeapPercent,
				diskPercent,
				node.DiskAvail,
				diskTotal,
				docsStr,
				heapUsed,
				heapMax,
				node.IP,
				name,
			}
			rows = append(rows, row)
		}

		formatter := ui.NewGenericTableFormatter()
		fmt.Print(formatter.FormatTable(headers, rows))
		fmt.Printf("Total: %d nodes\n", len(nodes))
	},
}

func init() {
	core.RootCmd.AddCommand(nodeCmd)
	nodeCmd.AddCommand(nodeDistCmd)
}
