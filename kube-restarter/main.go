package main

import (
	"fmt"
	"time"
)

const (
	RESTART_INTERVAL = 3
)

func main() {
	client, err := NewClient("/Users/h0n9/.kube/config")
	if err != nil {
		panic(err)
	}

	dps, err := client.GetDeployments("")
	if err != nil {
		panic(err)
	}

	fmt.Println("deployments:")
	for _, dp := range dps {
		fmt.Println("  - name:", dp.Name)
		fmt.Println("    namespace:", dp.Namespace)
		fmt.Println("    restarting:", true)
		_, err = client.RestartDeployment(&dp)
		if err != nil {
			fmt.Println("    restarted:", false)
			fmt.Println("    error:", err)
			continue
		}
		time.Sleep(RESTART_INTERVAL * time.Second)
		fmt.Println("    restarted:", true)
	}

	ss, err := client.GetStatefulSet("")
	if err != nil {
		panic(err)
	}

	fmt.Println("statefulset:")
	for _, s := range ss {
		fmt.Println("  - name:", s.Name)
		fmt.Println("    namespace:", s.Namespace)
		fmt.Println("    restarting:", true)
		_, err = client.RestartStatefulSet(&s)
		if err != nil {
			fmt.Println("    restarted:", false)
			fmt.Println("    error:", err)
			continue
		}
		time.Sleep(RESTART_INTERVAL * time.Second)
		fmt.Println("    restarted:", true)
	}
}
