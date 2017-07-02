package main

import(
	"net/http"
	"strings"
	"strconv"
	"errors"
)

var defaultChainSize = 1000

//This object should be created as a global variable, so as to pass around the entire
//program
type SessionManager struct {
	//chains is a slice of chain(chain is []session structs). This will be where all
	//the sessions are stored
	chains []chain
	
	//will be equivalent to indexing of chains (so 0 when there i 1 chain). Holds the
	//amount of chains - 1 (so as to start at 0)
	currentChains int 
	
	//The size of the each chain. If calling the default initializer, will be 
	//defaultChainSize (global)
	chainSize int
	
	//the placement of the earliest-end spot. Will be the beginning of the 
	//strung-together empty spots of ended sessions. Used by sm.nextSpot() and EndSession
	openSpot SpotMarker
	
	//The most-recently ended session spot. Everytime a session is ended, this spot gets
	//strung the newly-ended session spot, them lastEndedSpot is set to that newly-ended
	//spot. This and openSpot simulate a linked list of open places that will be used
	//first when StartSession calls nextSpot to start a session
	lastEndedSpot SpotMarker
	
	//the farthest placed spot. If there is empty spot from ended sessions, then 
	//StartSession will recieve this spot from nextSpot
	farthestPlaced SpotMarker

	//so that its not tried to be resized twice
	beingResized bool

	//when to resize
	resizeAt int

}

//a chains in SessionManager is []chain
type chain []session

//session structure
type session struct {
	active bool
	nextSpot SpotMarker
	
	identifier string	
}

//which chain its in, and which index
type SpotMarker struct {
	chain int
	index int	
}

/*******************************external functions for use*******************************/

/****************************initializers*****************************/
//default initializer
func NewSessionManager() *SessionManager{
	var sm SessionManager
	
	sm.chains = make([]chain, 1)
	sm.chainSize = defaultChainSize
	sm.chains[0] = make(chain, defaultChainSize)
	sm.currentChains = 0
	
	
	sm.openSpot = SpotMarker{-1, 0}
	sm.farthestPlaced = SpotMarker{0, -1}
	sm.lastEndedSpot = SpotMarker{}

	sm.beingResized = false
	sm.resizeAt = (defaultChainSize / 4) * 3

	return &sm
}

//will eventually have more customizable options, line initial amount of chains, when 
//to resize, how many chains to add on resize, etc. 
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

/*********************verify sessions stuff**************************/

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

/************************start session stuff*************************/

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

/*************************end session stuff****************************/

func (sm *SessionManager) EndSession(spot SpotMarker, identifier string)error{
	err := sm.VerifySession(spot, identifier)
	if err != nil {
		return errors.New("Trying to end invalid session: " + err.Error())
	}
	//lock
	sm.chains[spot.chain][spot.index].active = false
	if sm.openSpot.chain == -1 {
		sm.openSpot = spot
		sm.chains[spot.chain][spot.index].nextSpot.chain = -1
		sm.lastEndedSpot = spot
		return nil
	}
	sm.chains[sm.lastEndedSpot.chain][sm.lastEndedSpot.index].nextSpot = spot
	sm.lastEndedSpot = spot
	//unlock
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

/***********************cookie parsing*******************************/

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
	if sm.openSpot.chain == -1 {
		if sm.farthestPlaced.index <= sm.chainSize - 2{
			sm.farthestPlaced.index++
		} else {
			sm.farthestPlaced.index = 0
			sm.farthestPlaced.chain++
		}
		return sm.farthestPlaced
	}
	//lock
	spot := sm.openSpot
	if sm.lastEndedSpot == spot{
		sm.openSpot.chain = -1
	} else {
		sm.openSpot = sm.chains[spot.chain][spot.index].nextSpot
	}
	//unlock
	return spot
}

func (sm *SessionManager) resize(){
	sm.currentChains++
	newChains := make([]chain, sm.currentChains+1)
	//lock
	for i, val := range sm.chains {
		newChains[i] = val	
	}
	sm.chains = newChains
	sm.chains[1] = make(chain, sm.chainSize)
	//unlock
	sm.beingResized = false	
}

func newCookie(spot SpotMarker, identifier string) *http.Cookie{
	val := strconv.Itoa(spot.chain) + "|" + strconv.Itoa(spot.index) + "|" + identifier
	c := &http.Cookie{Name: "session", Value: val}
	return c
}







