package testshell

import (
	"fmt"
	"process"
	"resource"
	"strconv"
	"strings"
)

var (
	pm *process.ProcessManagement
)

func showCurPCB() {
	if pm.PCBCur == nil {
		fmt.Printf("* No Process is running\n")
	} else {
		fmt.Printf("* Process %s is running\n", pm.PCBCur.PID)
	}
}

func createProcess(args []string) {
	if i, err := strconv.Atoi(args[2]); err == nil {
		if err1 := pm.Create(process.PIDType(args[1]), i); err1 != nil {
			fmt.Println(err1)
		}
	} else {
		fmt.Printf("Incorrect parameter format：%s must be Integer\n", args[2])
	}
}

func deleteProcess(args []string) {
	if err1 := pm.Destroy(process.PIDType(args[1])); err1 != nil {
		fmt.Println(err1)
	}
}

func requestRes(args []string) {
	if i, err := strconv.Atoi(args[2]); err == nil {
		if err1 := pm.Request(resource.RIDType(args[1]), i); err1 != nil {
			fmt.Println(err1)
		}
	} else {
		fmt.Printf("Incorrect parameter format：%s must be Integer\n", args[2])
	}
}

func releaseRes(args []string) {
	if i, err := strconv.Atoi(args[2]); err == nil {
		if err1 := pm.Release(resource.RIDType(args[1]), i); err1 != nil {
			fmt.Println(err1)
		}
	} else {
		fmt.Printf("Incorrect parameter format：%s must be Integer\n", args[2])
	}
}

func timeout() {
	if err1 := pm.Timeout(); err1 != nil {
		fmt.Println(err1)
	}
}

func showPCBDetail(pcb *process.PCB, isLastChilds []bool) {
	fmt.Printf("%s\t%s\t",
		pcb.PID,
		pcb.Status,
	)
	if pcb.Parent == nil {
		fmt.Print("None\t")
	} else {
		fmt.Printf("%s\t", pcb.Parent.PID)
	}

	fmt.Printf("%d\t\t", pcb.Priority)

	if pcb.BlockingForRes == nil {
		fmt.Print("None\t")
	} else {
		fmt.Printf("%s:%d\t",
			pcb.BlockingForRes.Res.RID,
			pcb.BlockingForRes.N)
	}

	if pcb.OtherResource.Len() == 0 {
		fmt.Print("None\t")
	}
	for ele := pcb.OtherResource.Front(); ele != nil; ele = ele.Next() {
		pcbRes := ele.Value.(*process.PCBResource)
		fmt.Printf("%s:%d", pcbRes.Res.RID, pcbRes.N)
		if ele.Next() != nil {
			fmt.Print(";")
		}
	}

	if pcb.Children.Len() == 0 {
		fmt.Print("None\t")
	}
	for ele := pcb.Children.Front(); ele != nil; ele = ele.Next() {
		fmt.Printf("%s", ele.Value.(*process.PCB).PID)
		if ele.Next() != nil {
			fmt.Print(";")
		}
	}
	fmt.Println()
}

func status() {
	showCurPCB()
	fmt.Println()
	fmt.Println("* Process Detail Info")
	fmt.Println("PID\tStatus\tParent\tPriority\tWaitRes\tHoldRes\tChildren")
	traversePCBTree(pm.PCBRoot,
		make([]bool, 0),
		showPCBDetail)

	fmt.Println()
	fmt.Println("* Resource Detail Info")
	resList()
}

func traversePCBTree(pcb *process.PCB,
	isLastChilds []bool,
	handle func(pcb *process.PCB, isLastChilds []bool)) {

	handle(pcb, isLastChilds)

	for ele := pcb.Children.Front(); ele != nil; ele = ele.Next() {
		if ele.Next() == nil {
			traversePCBTree(ele.Value.(*process.PCB),
				append(isLastChilds, true),
				handle)
		} else {
			traversePCBTree(ele.Value.(*process.PCB),
				append(isLastChilds, false),
				handle)
		}
	}

}

func showPCBTree(pcb *process.PCB, isLastChilds []bool) {
	for i := 0; i < len(isLastChilds)-1; i++ {
		if isLastChilds[i] {
			fmt.Print("    ")
		} else {
			fmt.Print("│   ")
		}
	}
	if len(isLastChilds) >= 1 {
		if isLastChilds[len(isLastChilds)-1] {
			fmt.Print("└── ")
		} else {
			fmt.Print("├── ")
		}
	}
	fmt.Printf("%s\n", pcb.PID)
}

func pstree() {
	if pm.PCBRoot == nil {
		fmt.Println("No Process")
	} else {
		traversePCBTree(pm.PCBRoot,
			make([]bool, 0),
			showPCBTree)
	}

}

func resList() {
	fmt.Println("RID\tInit\tFree\tBlocking")
	for k, v := range pm.ResourceMap {
		fmt.Printf("%s\t%d\t%d\t", k, v.MaxResCnt, v.FreeResCnt)
		for ele := v.WaitingList.Front(); ele != nil; ele = ele.Next() {
			pcb := ele.Value.(*process.PCB)
			fmt.Print(pcb.PID, ":", pcb.BlockingForRes.N, " ")
		}
		if v.WaitingList.Len() == 0 {
			fmt.Print("None")
		}
		fmt.Println()
	}
}

//返回true表示接收到退出命令
func Exec(line string) bool {

	args := strings.Fields(line)

	switch {
	case len(args) == 1 && args[0] == "exit":
		return true
	case len(args) == 3 && args[0] == "cr":
		createProcess(args)
	case len(args) == 2 && args[0] == "de":
		deleteProcess(args)
	case len(args) == 3 && args[0] == "req":
		requestRes(args)
	case len(args) == 3 && args[0] == "rel":
		releaseRes(args)
	case len(args) == 1 && args[0] == "to":
		timeout()
	case len(args) == 1 && args[0] == "pstree":
		pstree()
		return false
	case len(args) == 1 && args[0] == "reslist":
		resList()
		return false
	case len(args) == 1 && args[0] == "status":
		status()
		return false
	default:
		fmt.Printf("未定义的命令：%s\n", line)
		return false
	}
	showCurPCB()
	return false
}

func init() {
	pm = process.NewPM()
	pm.AddResource(resource.NewRCB("R1", 1))
	pm.AddResource(resource.NewRCB("R2", 2))
	pm.AddResource(resource.NewRCB("R3", 3))
	pm.AddResource(resource.NewRCB("R4", 4))
	pm.Create("init", 0)
	showCurPCB()
}
