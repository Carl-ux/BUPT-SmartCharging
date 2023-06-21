package service

import (
	"BSC/model"
	"BSC/service"
	"errors"
	"fmt"
)

// 充电桩调度器
type PileScheduler struct {
	WaitingQueue     map[string]*ChargingRequest
	ExecutingRequest *ChargingRequest
	PileType         PileType
	IsBroken         bool
}

func NewPileScheduler(pileType PileType, status model.PileStatus) *PileScheduler {
	var brokenStatus bool
	if status == model.Running {
		brokenStatus = false
	} else {
		brokenStatus = true
	}
	return &PileScheduler{
		WaitingQueue:     make(map[string]*ChargingRequest),
		ExecutingRequest: nil,
		PileType:         pileType,
		IsBroken:         brokenStatus,
	}
}

func (p *PileScheduler) GetType() PileType {
	return p.PileType
}

func (p *PileScheduler) GetExecutingRequest() *ChargingRequest {
	return p.ExecutingRequest
}

func (p *PileScheduler) NextRequest() {
	if p.ExecutingRequest != nil {
		delete(p.WaitingQueue, p.ExecutingRequest.RequestID)
		p.ExecutingRequest = nil
	}

	if len(p.WaitingQueue) > 0 {
		for _, request := range p.WaitingQueue {
			request.IsExecuting = true
			request.BeginTime = service.GetDatetimeNow()
			p.ExecutingRequest = request
			fmt.Printf("充电桩 %d 开始执行请求 %s\n", p.ExecutingRequest.PileID, request.RequestID)
			break
		}
	}
}

func (p *PileScheduler) PushToQueue(request *ChargingRequest) {

	p.WaitingQueue[request.RequestID] = request
	if p.ExecutingRequest == nil {
		p.NextRequest()
	}
}

func (p *PileScheduler) GetUsedSize() int {
	return len(p.WaitingQueue)
}

func (p *PileScheduler) EstimateTime() int {
	totalCostTime := 0.0
	for _, request := range p.WaitingQueue {
		var power float64
		if request.RequestType == TricklePile {
			power = TricklePilePower
		} else {
			power = FastChargePilePower
		}
		cost := request.Amount / power * 3600
		totalCostTime += cost
	}
	return int(totalCostTime)
}

// 查看是否在等待队列中
func (p *PileScheduler) Contains(requestId string) bool {
	_, ok := p.WaitingQueue[requestId]
	return ok
}

// 移除请求
func (p *PileScheduler) Remove(requestId string) error {
	request, ok := p.WaitingQueue[requestId]
	if !ok {
		return errors.New("request_id not found in WaitingQueue")
	}
	if request.IsExecuting {
		p.NextRequest()
		return nil
	}
	delete(p.WaitingQueue, requestId)
	return nil
}

// 查找请求位置
func (p *PileScheduler) FindPosition(requestId string) (int, error) {
	posCnt := 0
	for reqId := range p.WaitingQueue {
		if reqId == requestId {
			return posCnt, nil
		}
		posCnt++
	}
	return -1, errors.New("request_id not found in WaitingQueue")
}

// 清除等待队列， includeExecuting为true时，清除正在执行的请求
func (p *PileScheduler) FetchAndClear(includeExecuting bool) []*ChargingRequest {
	var requests []*ChargingRequest
	for _, request := range p.WaitingQueue {
		requests = append(requests, request)
	}

	p.WaitingQueue = make(map[string]*ChargingRequest)

	if includeExecuting {
		p.ExecutingRequest = nil
	} else {
		if p.ExecutingRequest != nil {
			//清除正在执行的请求并将其重新放回等待队列
			//delete(p.WaitingQueue, p.ExecutingRequest.RequestID)
			p.WaitingQueue[p.ExecutingRequest.RequestID] = p.ExecutingRequest
			//删掉requests中的ExecutingRequest
			for i, request := range requests {
				if request.RequestID == p.ExecutingRequest.RequestID {
					requests = append(requests[:i], requests[i+1:]...)
					break
				}
			}
		}
	}
	//返回一个ChargingRequest类型的切片，其中包含所有从等待队列中提取的充电请求
	return requests
}
