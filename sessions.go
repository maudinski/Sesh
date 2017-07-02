package main

import(
	"net/http"
	"strings"
	"strconv"
	"errors"
)

//TODO make custom error types, for easy debugging for user

type SessionManager struct {
	
	chains []chain
	currentChains int //will be equivalent to indexing of array (so 0 when there i 1 chain)
	chainSize int

	openSpot SpotMarker
	farthestPlaced SpotMarker
	lastEndedSpot SpotMarker

	beingResized bool
	resizeAt int

}

type chain []session

type session struct {
	active bool
	nextSpot SpotMarker
	
	identifier string	
}

type SpotMarker struct {
	chain int
	index int	
}

//should be called before http.ListenAndServe
func NewCustomSessionManager(chainSize int) *SessionManager{
	var sm SessionManager
	
	sm.chains = make([]chain, 1)
	sm.chainSize = chainSize
	sm.chains[0] = make(chain, chainSize)
	sm.currentChains = 0
	
	
	sm.openSpot = SpotMarker{-1, 0}
	sm.farthestPlaced = SpotMarker{0, -1}
	sm.lastEndedSpot = SpotMarker{}

	sm.beingResized = false
	sm.resizeAt = (chainSize / 4) * 3

	return &sm
}

func NewSessionManager() *SessionManager{
	var sm SessionManager
	
	sm.chains = make([]chain, 1)
	sm.chainSize = 1000
	sm.chains[0] = make(chain, 1000)
	sm.currentChains = 0
	
	
	sm.openSpot = SpotMarker{-1, 0}
	sm.farthestPlaced = SpotMarker{0, -1}
	sm.lastEndedSpot = SpotMarker{}

	sm.beingResized = false
	sm.resizeAt = (1000 / 4) * 3

	return &sm
}


/****************************verify sessions stuff*********************************/

func (sm *SessionManager) VerifySession(spot SpotMarker, identifier string) error{	
	if spot.chain > sm.currentChains || spot.chain < 0 {
		return errors.New("Cookie has invalid chain index of " + strconv.Itoa(spot.chain))
	}
	if spot.index >= sm.chainSize || spot.index < 0 {
		return errors.New("Cookie has invalid sessin ind of "+strconv.Itoa(spot.index))	
	}
	if sm.chains[spot.chain][spot.index].identifier != identifier {
		return errors.New("Identfiers dont match")	
	}
	//maybe
	if !sm.chains[spot.chain][spot.index].active {
		return errors.New("Session is not active")
	}
	return nil
}

func (sm *SessionManager) VerifySessionCookie(c *http.Cookie) error {	
	spot, identifier, err := sm.ParseCookie(c)
	if err != nil{
		return err
	}
	return sm.VerifySession(spot, identifier)
}

/*******************************start session stuff********************************/

func (sm *SessionManager) StartSession(identifier string) SpotMarker {	
	spot := sm.nextSpot()
	sm.chains[spot.chain][spot.index] = session{true, SpotMarker{-1, 0}, identifier}
	go sm.checkResize(spot)
	return spot
}

func (sm *SessionManager) StartSessionCookie(identifier string) *http.Cookie{
	spot := sm.StartSession(identifier)
	return newCookie(spot, identifier)
}

/********************************end session stuff***********************************/

func (sm *SessionManager) EndSession(spot SpotMarker, identifier string)error{
	err := sm.VerifySession(spot, identifier)
	if err != nil {
		return errors.New("Trying to end invalid session: " + err.Error())
	}
	//lock some shit up
	sm.chains[spot.chain][spot.index].active = false
	if sm.openSpot.chain == -1 {
		sm.openSpot = spot
		sm.chains[spot.chain][spot.index].nextSpot.chain = -1
		sm.lastEndedSpot = spot
		return nil
	}
	sm.chains[sm.lastEndedSpot.chain][sm.lastEndedSpot.index].nextSpot = spot
	sm.lastEndedSpot = spot
	//unlock some shit up
	return nil
}


func (sm *SessionManager) EndSessionCookie(c *http.Cookie) error {
	spot, identifier, err := sm.ParseCookie(c)
	if err != nil {
		return err	
	}
	sm.EndSession(spot, identifier)
	return nil
}

/******************************cookie paring**************************************/

func (sm *SessionManager) ParseCookie(c *http.Cookie) (SpotMarker, string, error){
	var spot SpotMarker
	parts := strings.Split(c.Value, "|")	
	if len(parts) != 3{
		return spot, "", errors.New("Cookie Value isn't valid")	
	}
	chain, err := strconv.Atoi(parts[0])
	if err != nil{
		return spot, "", errors.New("Cookie Value isn't valid")	
	}
	index, err := strconv.Atoi(parts[1])
	if err != nil{
		return spot, "", errors.New("Cookie Value isn't valid")	
	}
	spot = SpotMarker{chain: chain, index: index}
	return spot, parts[2], nil
}

/**************************************internals****************************************/

func (sm *SessionManager) checkResize(spot SpotMarker){
	if sm.beingResized{
		return	
	}
	if spot.index >= sm.resizeAt && spot.chain == sm.currentChains {
		sm.beingResized = true
		sm.resize()	
	}
}

func (sm *SessionManager) nextSpot() SpotMarker {
	if sm.openSpot.chain == -1 { // or !sm.openSpots
		if sm.farthestPlaced.index <= sm.chainSize - 2{
			sm.farthestPlaced.index++
		} else {
			sm.farthestPlaced.index = 0
			sm.farthestPlaced.chain++
		}
		return sm.farthestPlaced
	}
	//lock some shit up
	spot := sm.openSpot
	if sm.lastEndedSpot == spot{
		sm.openSpot.chain = -1
	} else {
		sm.openSpot = sm.chains[spot.chain][spot.index].nextSpot
	}
	//unlock some shit up
	return spot
}

func (sm *SessionManager) resize(){
	sm.currentChains++
	newChains := make([]chain, sm.currentChains+1)
	//lock some shit up
	for i, val := range sm.chains {
		newChains[i] = val	
	}
	sm.chains = newChains
	sm.chains[1] = make(chain, sm.chainSize)
	//unlock some shit up
	sm.beingResized = false	
}

func newCookie(spot SpotMarker, identifier string) *http.Cookie{
	val := strconv.Itoa(spot.chain) + "|" + strconv.Itoa(spot.index) + "|" + identifier
	c := &http.Cookie{Name: "session", Value: val}
	return c
}







