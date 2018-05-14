package process

import (
	"container/list"
	"errors"
	"fmt"
	. "resource"
)

const MAXPriority = 3 //优先级数目

type ProcessManager interface {
	AddResource(rcb *RCB)                   //添加资源
	Create(PID PIDType, Priority int) error //创建进程
	Destroy(PID PIDType) error              //销毁进程
	Request(RID RIDType, n int) error       //请求资源
	Release(RID RIDType, n int) error       //释放资源
	Timeout() error                         //时间片轮转
	Schedule() error                        //调度
}

//进程管理
type ProcessManagement struct {
	PCBCur      *PCB                    //当前正在运行的PCB
	PCBRoot     *PCB                    //起始进程
	ReadyList   [MAXPriority]*list.List //准备队列，每个优先级一个
	ResourceMap map[RIDType]*RCB        //该机器所有的资源
	PCBMap      map[PIDType]*PCB        //通过名字查找PCB
}

//ProcessManagement构造函数
func NewPM() *ProcessManagement {
	pm := new(ProcessManagement)
	for i := range pm.ReadyList {
		pm.ReadyList[i] = list.New()
	}
	pm.ResourceMap = make(map[RIDType]*RCB)
	pm.PCBMap = make(map[PIDType]*PCB)
	return pm
}

//向准备队列中添加一个PCB
func (pm *ProcessManagement) addToReadyList(Priority int, pcb *PCB) {
	pm.ReadyList[Priority].PushBack(pcb)
}

func (pm *ProcessManagement) AddResource(rcb *RCB) {
	var rid RIDType = rcb.RID
	pm.ResourceMap[rid] = rcb
}

func (pm *ProcessManagement) Create(PID PIDType, Priority int) error {
	if Priority >= MAXPriority {
		return errors.New(fmt.Sprintf("该优先级大于最高优先级：%d", (MAXPriority - 1)))
	}
	if _, ok := pm.PCBMap[PID]; ok {
		return errors.New(fmt.Sprintf("该进程已存在"))
	}

	//创建一个PCB
	pcb := NewPCB(PID, Priority)

	//该进程为根进城
	if pm.PCBRoot == nil {
		pm.PCBRoot = pcb
	}

	//构造进程树
	if pm.PCBCur != nil { //当前不是第一个进程
		pcb.SetParent(pm.PCBCur)
		pm.PCBCur.AddChild(pcb)
	}
	pm.addToReadyList(Priority, pcb) //添加到等待队列
	pm.PCBMap[PID] = pcb             //添加到map中
	pm.Schedule()                    //调度

	return nil
}

func (pm *ProcessManagement) DestroyByPCB(pcb *PCB) {
	//删除的进程是正在运行的进程
	if pcb == pm.PCBCur {
		pm.PCBCur = nil
	}
	//设为不可用（也就是删除，真正的删除在调度的时候）
	pcb.Disable = true
	delete(pm.PCBMap, pcb.PID) //从map中删除

	//释放占用的资源
	for ele := pcb.OtherResource.Front(); ele != nil; ele = ele.Next() {
		pcbRes := ele.Value.(*PCBResource)
		pcb.Release(pcbRes.Res, pcbRes.N) //释放全部资源
		pm.ScanBlockForRes(pcbRes.Res)    //扫描阻塞在该资源的进程
	}
	//如果有父亲，则删除进程关系
	if pcb.Parent != nil {
		pcb.Parent.DeleteChild(pcb.PID) //从父进程中删除孩子
		pcb.SetParent(nil)              //从当前进程删除父进程的引用
	}
	//递归删除子进程
	for ele := pcb.Children.Front(); ele != nil; {
		childPCB := ele.Value.(*PCB)
		ele = ele.Next()
		pm.DestroyByPCB(childPCB)
	}
}

func (pm *ProcessManagement) Destroy(PID PIDType) error {
	if pcb, ok := pm.PCBMap[PID]; ok {
		//若该进程为根进程，清空
		if pcb == pm.PCBRoot {
			pm.PCBRoot = nil
		}

		pm.DestroyByPCB(pcb)
		//重新调度
		pm.Schedule()
	} else {
		return errors.New(fmt.Sprintf("该进程不存在"))
	}
	return nil
}

func (pm *ProcessManagement) Request(RID RIDType, n int) error {
	if pm.PCBCur == nil {
		return errors.New("当前没有进程在执行")
	}
	if rcb, ok := pm.ResourceMap[RID]; ok {
		//请求资源
		if err := pm.PCBCur.Request(rcb, n); err != nil {
			return err
		}
		if pm.PCBCur.Status == PCB_BLOCKED { //进入了阻塞状态
			pm.PCBCur = nil      //清空当前正在执行的进程
			return pm.Schedule() //进行调度
		}
	} else {
		return errors.New("没有此资源")
	}
	return nil
}

//扫描等待在该rcb上的pcb是否可以唤醒
func (pm *ProcessManagement) ScanBlockForRes(rcb *RCB) bool {
	//检查该rcb是否可以给第一个阻塞的pcb使用
	//拿到第一个
	if ele := rcb.WaitingList.Front(); ele != nil {
		//检查资源是否足够使用
		pcb := ele.Value.(*PCB)
		//足够进行使用
		if pcb.BlockingForRes.N <= rcb.FreeResCnt {
			pcb.Request(rcb, pcb.BlockingForRes.N) //重新申请资源

			pcb.Status = PCB_READY               //进程变成就绪状态
			pm.addToReadyList(pcb.Priority, pcb) //加入就绪队列

			rcb.WaitingList.Remove(ele) //将PCB从RCB的阻塞列表删除
			pcb.BlockingForRes = nil    //清零阻塞等待的资源
			return true
		}
	}
	return false
}

func (pm *ProcessManagement) Release(RID RIDType, n int) error {
	if pm.PCBCur == nil {
		return errors.New("当前没有进程在执行")
	}
	if rcb, ok := pm.ResourceMap[RID]; ok {
		//释放资源
		pm.PCBCur.Release(rcb, n)

		if pm.ScanBlockForRes(rcb) {
			pm.Schedule() //重新调度
		}

	} else {
		return errors.New("没有此资源")
	}
	return nil
}

func (pm *ProcessManagement) Timeout() error {
	//当前没有进程在运行，直接调用调度
	if pm.PCBCur == nil {
		return pm.Schedule()
	}

	lastPCB := pm.PCBCur //记录下当前PCB
	pm.PCBCur = nil      //清除当前PCB

	//调度
	if err := pm.Schedule(); err != nil {
		return err
	}

	//没有可调度的，恢复
	if pm.PCBCur == nil {
		pm.SetRunning(lastPCB)
		return nil
	}
	//调度之后的PCB优先级低于之前PCB的优先级，
	//禁止调度，恢复
	if pm.PCBCur.Priority < lastPCB.Priority {
		//重新放回等待队列，注意是首部
		pm.ReadyList[pm.PCBCur.Priority].PushFront(pm.PCBCur)
		pm.SetRunning(lastPCB)
	} else { //放入就绪队列
		lastPCB.Status = PCB_READY
		pm.addToReadyList(lastPCB.Priority, lastPCB)
	}
	return nil
}

func (pm *ProcessManagement) SetRunning(pcb *PCB) {
	pm.PCBCur = pcb
	pm.PCBCur.Status = PCB_RUNNING
}

func (pm *ProcessManagement) Schedule() error {
	for i := MAXPriority - 1; i >= 0; i-- {
		if pm.PCBCur != nil && i <= pm.PCBCur.Priority {
			break
		}
		l := pm.ReadyList[i]
		if l.Len() == 0 {
			continue
		}
		nextEle := l.Front()
		nextVal := nextEle.Value.(*PCB)
		l.Remove(nextEle)
		//当前pcb不可用（已被删除），递归调用调度
		if nextVal.Disable {
			return pm.Schedule()
		}

		//放入就绪队列
		if pm.PCBCur != nil {
			pm.addToReadyList(pm.PCBCur.Priority, pm.PCBCur)
		}
		pm.SetRunning(nextVal)
	}
	return nil
}
