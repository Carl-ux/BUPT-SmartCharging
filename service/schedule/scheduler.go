package service

import (
	"BSC/dao"
	"BSC/model"
	"BSC/service"
	"BSC/utils"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"sync"
	"time"
)

var Schd *Scheduler

func InitSchd() {
	Schd = NewScheduler()

	// //创建初始请求
	// var testName = "test"
	// for i := 1; i <= 2; i++ {
	// 	testUserName := fmt.Sprintf("%s%d", testName, i)
	// 	Schd.SubmitRequest(TricklePile, testUserName, 14, 200, false)
	// }
	// for i := 3; i <= 4; i++ {
	// 	testUserName := fmt.Sprintf("%s%d", testName, i)
	// 	Schd.SubmitRequest(FastChargePile, testUserName, 30, 200, false)
	// }
}

// ID分配器结构体
type RequestIDAllocator struct {
	lock    sync.Mutex
	idFlags [2][MaxRecycleID]bool
	curID   [2]int
}

func NewRequestIdAllocator() *RequestIDAllocator {
	return &RequestIDAllocator{
		lock:    sync.Mutex{},
		idFlags: [2][MaxRecycleID]bool{},
		curID:   [2]int{0, 0},
	}
}

func (a *RequestIDAllocator) Alloc(requestMode int) (string, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.curID[requestMode] = (a.curID[requestMode] + 1) % MaxRecycleID

	a.idFlags[requestMode][a.curID[requestMode]] = true

	var requestID string
	switch requestMode {
	case int(TricklePile):
		requestID = "T" + strconv.Itoa(a.curID[requestMode])
	case int(FastChargePile):
		requestID = "F" + strconv.Itoa(a.curID[requestMode])
	}
	return requestID, nil
}

func (a *RequestIDAllocator) Dealloc(requestMode int, chargingID string) {
	a.lock.Lock()
	defer a.lock.Unlock()
	index, _ := strconv.Atoi(chargingID[1:])
	a.idFlags[requestMode][index] = false
}

type Scheduler struct {
	IdAllocator       *RequestIDAllocator
	Lock              sync.Mutex
	CheckLock         sync.Mutex
	PileSchedulers    map[int]*PileScheduler
	WaitingAreaMap    map[string]*ChargingRequest
	UsernameToRequest map[string]string
	SchedulingMode    SchedulingMode
	BrokenPileID      int
	RecoveryQueue     []*ChargingRequest
	WaitingAreas      map[PileType][]*ChargingRequest
}

func NewScheduler() *Scheduler {
	s := &Scheduler{
		IdAllocator:       NewRequestIdAllocator(),
		Lock:              sync.Mutex{},
		CheckLock:         sync.Mutex{},
		PileSchedulers:    make(map[int]*PileScheduler),
		WaitingAreaMap:    make(map[string]*ChargingRequest),
		UsernameToRequest: make(map[string]string),
		SchedulingMode:    Normal,
		WaitingAreas:      make(map[PileType][]*ChargingRequest),
	}

	s.WaitingAreas[TricklePile] = []*ChargingRequest{}
	s.WaitingAreas[FastChargePile] = []*ChargingRequest{}

	// Initialize pile schedulers.
	piles, _ := dao.GetAllPiles()
	for _, pile := range piles {
		log.Println(pile.PileId, pile.Type, pile.ChargerStatus)
		if pile.Type == "T" {
			s.PileSchedulers[pile.PileId] = NewPileScheduler(TricklePile, pile.ChargerStatus)
		}
		if pile.Type == "F" {
			s.PileSchedulers[pile.PileId] = NewPileScheduler(FastChargePile, pile.ChargerStatus)
		}
		if pile.ChargerStatus != model.Running {
			s.BrokenPileID = pile.PileId
		}
	}
	go s.checkProc()
	return s
}

// 将队列首的请求从队列中弹出
func (s *Scheduler) PopQueue(queue *[]*ChargingRequest) *ChargingRequest {
	for len(*queue) > 0 && (*queue)[0].IsRemoved {
		*queue = (*queue)[1:]
	}
	if len(*queue) == 0 {
		return nil
	}
	request := (*queue)[0]
	*queue = (*queue)[1:]
	return request
}

// 将请求压入队列
func (s *Scheduler) PushQueue(queue *[]*ChargingRequest, request *ChargingRequest) {
	*queue = append(*queue, request)
}

// 找到空闲充电桩
func (s *Scheduler) FindFastestSparePile(requestType PileType) int {
	fastestPile := -1
	shortestTime := float64(1<<63 - 1)

	for pileID, pileScheduler := range s.PileSchedulers {
		if pileScheduler.IsBroken {
			continue
		}
		if pileScheduler.GetType() != requestType {
			continue
		}
		if pileScheduler.GetUsedSize() == ChargingQueueLen {
			continue
		}
		//fmt.Printf("pileID: %d, usedsize: %v\n", pileID, pileScheduler.GetUsedSize())
		cost := float64(pileScheduler.EstimateTime())
		if shortestTime <= cost {
			continue
		}
		fastestPile = pileID
		shortestTime = cost
	}
	return fastestPile
}

// 找到恢复位置
func (s *Scheduler) FindRecoveryPosition(requestID string) int {
	pos := 0
	for _, request := range s.RecoveryQueue {
		if request.RequestID == requestID {
			return pos
		}
		pos++
	}

	return -1
}

// 调度
func (s *Scheduler) TrySchedule() {
	scheduleOnType := func(pileType PileType) {
		for {
			targetPile := s.FindFastestSparePile(pileType)
			//log.Print("fastestPile: ", targetPile)
			if targetPile == -1 {
				return
			}
			waitingAreaQueue := s.WaitingAreas[pileType]
			request := s.PopQueue(&waitingAreaQueue)
			if request == nil {
				return
			}
			s.WaitingAreas[pileType] = waitingAreaQueue
			request.IsInWaitingQueue = true
			request.PileID = targetPile
			s.PileSchedulers[targetPile].PushToQueue(request)
			fmt.Printf("[scheduler] request %s has been moved into queue of pile %d\n", request.RequestID, request.PileID)
		}
	}

	skipTypes := map[PileType]struct{}{}
	//跳过进行故障调度的充电桩
	if s.SchedulingMode != Normal {
		pileType := s.PileSchedulers[s.BrokenPileID].GetType()
		skipTypes[pileType] = struct{}{}

		for len(s.RecoveryQueue) > 0 {
			targetPile := s.FindFastestSparePile(pileType)
			//队列全满
			if targetPile == -1 {
				break
			}
			request := s.RecoveryQueue[0]
			s.RecoveryQueue = s.RecoveryQueue[1:]
			request.FailFlag = false
			request.PileID = targetPile
			s.PileSchedulers[targetPile].PushToQueue(request)
			fmt.Printf("[recovery] request %s has been moved into queue of pile %d\n", request.RequestID, targetPile)
			//故障队列调度完成
			if len(s.RecoveryQueue) == 0 {
				fmt.Println("[recovery] recovery queue is empty now. resume scheduling.")
				s.SchedulingMode = Normal
				for key := range skipTypes {
					//清空 恢复正常调度
					delete(skipTypes, key)
				}
			}
		}
	}
	//未被故障影响的充电桩进行调度
	if _, ok := skipTypes[TricklePile]; !ok {
		scheduleOnType(TricklePile)
	}
	if _, ok := skipTypes[FastChargePile]; !ok {
		scheduleOnType(FastChargePile)
	}
}

func (s *Scheduler) checkIfCompleted(request *ChargingRequest) bool {
	timeNow := service.GetDatetimeNow()
	var power float64
	if request.RequestType == TricklePile {
		power = TricklePilePower
	} else {
		power = FastChargePilePower
	}
	amount := request.Amount
	beginTime := request.BeginTime
	completeTime := beginTime.Add(time.Duration(float64(amount)/power*3600) * time.Second)
	return timeNow.After(completeTime)
}

func (s *Scheduler) calcChargedAmount(request *ChargingRequest, timeNow time.Time) float64 {
	var power float64
	if request.RequestType == TricklePile {
		power = TricklePilePower
	} else {
		power = FastChargePilePower
	}
	chargedAmount := (timeNow.Sub(request.BeginTime).Seconds() / 3600) * power
	return chargedAmount
}

func (s *Scheduler) CalcChargedAmount(request *ChargingRequest) float64 {
	var power float64
	if request.RequestType == TricklePile {
		power = TricklePilePower
	} else {
		power = FastChargePilePower
	}
	chargedTime := service.GetDatetimeNow().Sub(request.BeginTime).Seconds()
	return chargedTime / 3600 * power
}

func (s *Scheduler) checkProc() {
	for {
		s.CheckLock.Lock()
		for _, pileScheduler := range s.PileSchedulers {
			s.Lock.Lock()
			executingRequest := pileScheduler.GetExecutingRequest()
			if executingRequest == nil {
				s.Lock.Unlock()
				continue
			}
			isCompleted := s.checkIfCompleted(executingRequest)
			s.Lock.Unlock()

			if isCompleted {
				fmt.Printf("[scheduler] request %s completed.\n", executingRequest.RequestID)
				s.EndRequest(executingRequest.RequestID, service.GetDatetimeNow())
			}
		}
		s.CheckLock.Unlock()
		time.Sleep(1 * time.Second)
	}
}

func (s *Scheduler) SubmitRequest(requestMode PileType, username string, amount float64, batteryCapacity float64, requeue bool) error {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	if _, ok := s.UsernameToRequest[username]; ok {
		return errors.New("用户已经有请求在等候区")
	}

	usedSize := 0
	for _, queue := range s.WaitingAreas {
		for _, request := range queue {
			if !request.IsRemoved {
				usedSize++
			}
		}
	}

	waitingQueue := s.WaitingAreas[requestMode]

	if usedSize == WaitingAreaCapacity {
		return errors.New("等候区已满")
	}

	requestID, _ := s.IdAllocator.Alloc(int(requestMode))
	request := ChargingRequest{
		RequestID:       requestID,
		RequestType:     requestMode,
		Username:        username,
		Amount:          amount,
		BatteryCapacity: batteryCapacity,
		CreateTime:      service.GetDatetimeNow(),
	}

	//fmt.Println(request.RequestID, request.RequestType, request.Username, request.Amount, request.BatteryCapacity, request.CreateTime)
	request.RequeueFlag = requeue

	s.WaitingAreaMap[requestID] = &request
	s.UsernameToRequest[username] = requestID
	s.PushQueue(&waitingQueue, &request)
	s.WaitingAreas[requestMode] = waitingQueue
	fmt.Printf("[scheduler] request %s from user %s is submitted\n", requestID, username)

	// 等待区更新 尝试调度
	s.TrySchedule()
	return nil
}

func (s *Scheduler) UpdateRequest(requestID string, amount float64, requestType PileType) error {

	request := s.WaitingAreaMap[requestID]
	if request.IsInWaitingQueue {
		return errors.New("不允许在充电区更新请求")
	}

	if request.RequestType == requestType {
		request.Amount = amount
		return nil
	}

	// 修改了模式
	s.EndRequest(requestID, service.GetDatetimeNow())
	s.SubmitRequest(requestType,
		request.Username,
		amount,
		request.BatteryCapacity,
		true)

	return nil
}

func (s *Scheduler) EndRequest(requestID string, time time.Time) {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	request := s.WaitingAreaMap[requestID]
	requestType := request.RequestType
	request.IsRemoved = true
	delete(s.WaitingAreaMap, requestID)

	s.IdAllocator.Dealloc(int(requestType), requestID)
	delete(s.UsernameToRequest, request.Username)
	pileID := request.PileID
	pileScheduler := s.PileSchedulers[pileID]
	if request.IsInWaitingQueue {
		pileScheduler.Remove(requestID)
	}
	var chargedAmount float64
	if request.IsExecuting {
		if !s.checkIfCompleted(request) {
			fmt.Printf("[scheduler] request %s is cancelled while executing.\n", request.RequestID)
			chargedAmount = s.CalcChargedAmount(request)
		} else {
			chargedAmount = request.Amount
		}
		// 触发结算流程生成详单
		fmt.Printf("[scheduler] request %s created an order.\n", requestID)
		CreateOrder(
			requestType,
			pileID,
			request.Username,
			chargedAmount,
			request.BeginTime,
			time,
			request.FailFlag,
		)
	} else {
		fmt.Printf("[scheduler] request %s is cancelled.\n", requestID)
	}

	// pile_scheduler 有空位 触发调度流程
	s.TrySchedule()
}

func (s *Scheduler) GetRequestStatus(requestID string) RequestStatus {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	request := s.WaitingAreaMap[requestID]
	if request.IsRemoved {
		return RequestStatus{NotCharging, -1, 0}
	}
	if request.IsExecuting {
		return RequestStatus{Charging, 1, request.PileID}
	}
	if request.FailFlag {
		pos := s.FindRecoveryPosition(requestID)
		return RequestStatus{FailRequeue, pos, 0}
	} else if request.IsInWaitingQueue {
		pileID := request.PileID
		pileScheduler := s.PileSchedulers[pileID]
		pos, err := pileScheduler.FindPosition(requestID)
		if err != nil {
			fmt.Printf("[scheduler] request %s is in waiting queue but not found in pile scheduler.\n", requestID)
			return RequestStatus{NotCharging, -1, 0}
		}
		//haha
		if pos == 0 {
			pos = 1
		}

		return RequestStatus{WaitingStage2, pos, pileID}
	}
	status := WaitingStage1
	if request.RequeueFlag {
		status = ChangeModeRequeue
	}
	requestType := request.RequestType
	requestWaitingQueue := s.WaitingAreas[requestType]
	posCnt := 0
	for _, req := range requestWaitingQueue {
		if req.RequestID == requestID {
			break
		}
		if req.IsRemoved {
			continue
		}
		posCnt++
	}

	usedSize := 0
	for _, p := range s.PileSchedulers {
		if p.GetUsedSize() > usedSize {
			usedSize = p.GetUsedSize()
		}
	}
	posCnt += usedSize

	return RequestStatus{status, posCnt, 0}
}

func (s *Scheduler) GetRequestIDByUsername(username string) (string, error) {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	requestID, ok := s.UsernameToRequest[username]
	if !ok {
		return "0", fmt.Errorf("用户未创建充电请求")
	}
	return requestID, nil
}

// 故障
func (s *Scheduler) Brake(pileID int) {
	s.CheckLock.Lock()
	defer s.CheckLock.Unlock()

	// debug("[recovery] pile %d is down.", pileID)

	t := service.GetDatetimeNow()

	s.SchedulingMode = DefalutScheduleMode
	s.BrokenPileID = pileID
	pileScheduler := s.PileSchedulers[pileID]
	pileScheduler.IsBroken = true

	//
	requests := pileScheduler.FetchAndClear(true)
	fmt.Printf("[brake] request %s is cancelled while executing.\n", requests[0].RequestID)
	// 触发结算流程生成详单
	fmt.Printf("[brake] request %s created an order.\n", requests[0].RequestID)
	CreateOrder(
		requests[0].RequestType,
		pileID,
		requests[0].Username,
		s.calcChargedAmount(requests[0], t),
		requests[0].BeginTime,
		service.GetDatetimeNow(),
		requests[0].FailFlag,
	)

	requests[0].Amount = requests[0].Amount - s.calcChargedAmount(requests[0], t)
	requests[0].BeginTime = t

	// executingRequest := pileScheduler.GetExecutingRequest()
	// if executingRequest != nil {
	// 	executingRequest.FailFlag = true
	// 	s.EndRequest(executingRequest.RequestID)
	// }

	switch s.SchedulingMode {
	case Priority:

		for _, request := range requests {
			// debug("[recovery] request %d has been moved to recovery queue.", request.requestID)
			request.PileID = 0
			request.FailFlag = true
		}
		s.RecoveryQueue = requests

	//TODO FIFO调度，需要完善
	case TimeOrdered:
		pileType := pileScheduler.GetType()
		requests := []*ChargingRequest{}
		for PileID, scheduler := range s.PileSchedulers {
			if scheduler.GetType() != pileType {
				continue
			}
			includeExecuting := false
			if PileID == pileID {
				includeExecuting = true
			}
			reqs := scheduler.FetchAndClear(includeExecuting)
			requests = append(requests, reqs...)
		}
		for _, request := range requests {
			// debug("[recovery] request %d has been moved to recovery queue.", request.requestID)
			request.PileID = 0
			request.FailFlag = true
		}
		// 按照 requestID 排序
		sort.SliceStable(requests, func(i, j int) bool {
			return requests[i].RequestID < requests[j].RequestID
		})
		s.RecoveryQueue = requests
	}
	s.TrySchedule()
}

// TODO 故障恢复
// 恢复策略
func (s *Scheduler) Recover(pileID int) {
	s.CheckLock.Lock()
	defer s.CheckLock.Unlock()

	// debug("[recovery] pile %d is up.", pileID)

	s.SchedulingMode = Recovery
	pileScheduler := s.PileSchedulers[pileID]
	pileScheduler.IsBroken = false

	pileType := pileScheduler.GetType()
	requests := []*ChargingRequest{}
	for _, scheduler := range s.PileSchedulers {
		if scheduler.GetType() != pileType {
			continue
		}
		reqs := scheduler.FetchAndClear(false)
		requests = append(requests, reqs...)
	}
	for _, request := range requests {
		// debug("[recovery] request %d has been moved to recovery queue.", request.requestID)
		request.PileID = 0
		request.FailFlag = true
	}
	sort.SliceStable(requests, func(i, j int) bool {
		return requests[i].RequestID[1:] < requests[j].RequestID[1:]
	})
	s.RecoveryQueue = requests
	s.TrySchedule()
}

func (s *Scheduler) Snapshot() []map[string]interface{} {
	requestList := []map[string]interface{}{}

	//排序无序map键值对
	keys := make([]string, 0, len(s.WaitingAreaMap))
	for k := range s.WaitingAreaMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		request := s.WaitingAreaMap[k]
		requestInfo := map[string]interface{}{
			//int 转 string:  strconv.Itoa(int)
			"pile_id":        strconv.Itoa(request.PileID),
			"username":       request.Username,
			"battery_size":   request.BatteryCapacity,
			"require_amount": request.Amount,
			"waiting_time":   utils.Decimal(service.GetDatetimeNow().Sub(request.CreateTime).Seconds()),
		}
		requestList = append(requestList, requestInfo)
	}
	return requestList
}
