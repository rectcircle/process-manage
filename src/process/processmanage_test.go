package process

import (
	"fmt"
	. "resource"
	"testing"
)

func TestNewPM(t *testing.T) {
	// pm := NewPM()
	// fmt.Println(pm)
}

func TestCreate(t *testing.T) {
	pm := NewPM()
	pm.Create("init", 0)
	if pm.ReadyList[0].Len() != 0 || pm.ReadyList[1].Len() != 0 || pm.ReadyList[2].Len() != 0 {
		t.Errorf("init调度错误")
	}
	pm.Create("x", 1)
	if pm.PCBCur.PID != "x" {
		t.Error("x调度错误")
	}
	if pm.ReadyList[0].Len() != 1 || pm.ReadyList[1].Len() != 0 || pm.ReadyList[2].Len() != 0 {
		t.Errorf("错误")
	}

	// fmt.Println(pm)
	// fmt.Println(pm.PCBCur)
	// fmt.Println(pm.ReadyList[0].Len())
	// fmt.Println(pm.ReadyList[1].Len())
	// fmt.Println(pm.ReadyList[2].Len())
}

func TestTimeout(t *testing.T) {
	pm := NewPM()
	pm.Create("init", 0)
	pm.Create("p", 1)
	if pm.PCBCur.PID != "p" {
		t.Error("p调度错误")
	}
	pm.Create("q", 1)
	if pm.PCBCur.PID != "p" {
		t.Error("q调度错误")
	}

	pm.Timeout()
	if pm.PCBCur.PID != "q" {
		t.Error("Timeout调度错误")
	}

	pm.Timeout()
	if pm.PCBCur.PID != "p" {
		t.Error("Timeout调度错误")
	}
}

func TestDestroy(t *testing.T) {
	pm := NewPM()
	pm.Create("init", 0)
	pm.Create("p", 1)
	pm.Create("q", 1)
	pm.Destroy("p")

	if pm.PCBCur.PID != "init" {
		t.Error("Destroy错误")
	}

	if pm.ReadyList[0].Len() != 0 || pm.ReadyList[1].Len() != 0 || pm.ReadyList[2].Len() != 0 {
		t.Errorf("Destroy调度错误")
	}
}

func TestAll(t *testing.T) {
	pm := NewPM()
	pm.AddResource(NewRCB("R1", 1))
	pm.AddResource(NewRCB("R2", 2))
	pm.AddResource(NewRCB("R3", 3))
	pm.AddResource(NewRCB("R4", 4))
	pm.Create("init", 0)
	fmt.Println(pm.PCBCur.PID)
	pm.Create("x", 1)
	fmt.Println(pm.PCBCur.PID)
	pm.Create("p", 1)
	fmt.Println(pm.PCBCur.PID)
	pm.Create("q", 1)
	fmt.Println(pm.PCBCur.PID)
	pm.Create("r", 1)
	fmt.Println(pm.PCBCur.PID)
	fmt.Println(pm.PCBCur.Children.Len())
	pm.Timeout()
	fmt.Println(pm.PCBCur.PID)
	pm.Request("R2", 1)
	fmt.Println(pm.PCBCur.PID)
	pm.Timeout()
	fmt.Println(pm.PCBCur.PID)
	pm.Request("R3", 3)
	fmt.Println(pm.PCBCur.PID)
	pm.Timeout()
	fmt.Println(pm.PCBCur.PID)
	pm.Request("R4", 3)
	fmt.Println(pm.PCBCur.PID)
	pm.Timeout()
	fmt.Println(pm.PCBCur.PID)
	pm.Timeout()
	fmt.Println(pm.PCBCur.PID)
	pm.Request("R3", 1)
	fmt.Println(pm.PCBCur.PID)
	pm.Request("R4", 2)
	fmt.Println(pm.PCBCur.PID)
	pm.Request("R2", 2)
	fmt.Println(pm.PCBCur.PID)
	pm.Timeout()
	fmt.Println(pm.PCBCur.PID)
	pm.Destroy("q")
	fmt.Println(pm.PCBCur.PID)
	pm.Timeout()
	fmt.Println(pm.PCBCur.PID)
	pm.Timeout()
	fmt.Println(pm.PCBCur.PID)
	/*
		init
		cr x 1
		cr p 1
		cr q 1
		cr r 1
		to
		req R2 1
		to
		req R3 3
		to
		req R4 3
		to
		to
		req R3 1
		req R4 2
		req R2 2
		to
		de q
		to
		to
	*/
}
