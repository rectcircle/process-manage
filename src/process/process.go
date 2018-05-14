package process

import (
	"container/list"
	"errors"
	. "resource"
)

type PIDType string

// type RIDType resource.RIDType
// type RCB resource.RCB
type StatusType string

type PCBResource struct {
	Res *RCB //指向资源的指针
	N   int
}

//PCB状态
const (
	PCB_READY   = "ready"
	PCB_RUNNING = "running"
	PCB_BLOCKED = "blocked"
)

type PCB struct {
	PID            PIDType      //进程id
	OtherResource  *list.List   //进程占有的资源
	Status         StatusType   //进程状态
	BlockingForRes *PCBResource //阻塞等待的RCB
	Parent         *PCB         //构成进程树
	Children       *list.List   //构成进程树
	Priority       int          //0, 1, 2 (Init, User, System)
	Disable        bool         //当前tcb是否被注销，默认为false
}

//创建一个就绪状态的PCB
func NewPCB(PID PIDType, Priority int) *PCB {
	pcb := &PCB{}
	pcb.PID = PID
	pcb.Priority = Priority
	pcb.Status = PCB_READY
	pcb.Children = list.New()
	pcb.OtherResource = list.New()
	return pcb
}

func (pcb *PCB) SetParent(parent *PCB) {
	pcb.Parent = parent
}

func (pcb *PCB) AddChild(child *PCB) {
	pcb.Children.PushBack(child)
}

func (pcb *PCB) DeleteChild(PID PIDType) {
	for ele := pcb.Children.Front(); ele != nil; ele = ele.Next() {
		cPCB := ele.Value.(*PCB)
		if cPCB.PID == PID {
			pcb.Children.Remove(ele)
			break
		}
	}
}

func (pcb *PCB) FindPCBRes(rcb *RCB) *list.Element {
	for ele := pcb.OtherResource.Front(); ele != nil; ele = ele.Next() {
		if pcbRes := ele.Value.(*PCBResource); pcbRes.Res == rcb {
			return ele
		}
	}
	return nil
}

func (pcb *PCB) Request(rcb *RCB, n int) error {
	if n > rcb.MaxResCnt {
		return errors.New("申请资源数大于该资源最大资源数")
	}
	pcbRes := &PCBResource{rcb, n}
	if n > rcb.FreeResCnt { //资源不够申请
		pcb.Status = PCB_BLOCKED      //进程阻塞
		rcb.WaitingList.PushBack(pcb) //将PCB加入RCB的阻塞列表
		pcb.BlockingForRes = pcbRes   //设置阻塞等待的资源
	} else { //资源足够申请
		rcb.FreeResCnt -= n //更新资源
		ele := pcb.FindPCBRes(rcb)
		if ele == nil {
			pcb.OtherResource.PushBack(pcbRes) //持有资源
		} else {
			pcbRes := ele.Value.(*PCBResource)
			pcbRes.N += n
		}
	}
	return nil
}

func (pcb *PCB) Release(rcb *RCB, n int) {
	ele := pcb.FindPCBRes(rcb)
	pcbRes := ele.Value.(*PCBResource)
	if n > pcbRes.N { //要求释放的资源大于持有的资源，释放全部
		n = pcbRes.N
	}
	//更新持有资源
	pcbRes.N -= n
	//释放资源
	pcbRes.Res.FreeResCnt += n
	if pcbRes.N <= 0 {
		pcb.OtherResource.Remove(ele) //从资源列表删除
	}
}
