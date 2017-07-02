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

//custom initializer
//will eventually have more customizable options, like initial amount of chains, when 
//to resize, how many chains to add on resize, http.Cookie Name & maybe Value format, etc
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

//takes in the spot of the session and the identifier
//
//The first 2 if statements check to make sure that the cookie's values werent 
//tampered with (to prevent against someone tampering with their cookies and
//trying to break the server). 3rd and 4th if statement check if the identifier
//matches and if the session is active, respectively. Returns nil if the session is valid
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

//takes in the cookie. Parses the cookie and returns sm.VerifySession
func (sm *SessionManager) VerifySessionCookie(c *http.Cookie) error {	
	spot, identifier, err := sm.ParseCookie(c)
	if err != nil{
		return err
	}
	return sm.VerifySession(spot, identifier)
}

/************************start session stuff*************************/

//takes the identifier for the session
//requests the next spot from sm.nextSpot(), sets the sessions in sm.chains, and runs 
//concurrently sm.checkResize. returns the spot
func (sm *SessionManager) StartSession(identifier string) SpotMarker {	
	spot := sm.nextSpot()
	sm.chains[spot.chain][spot.index] = session{true, SpotMarker{-1, 0}, identifier}
	go sm.checkResize(spot)
	return spot
}

//takes the identifier string. calls sm.StartSession, then returns a new cookie 
func (sm *SessionManager) StartSessionCookie(identifier string) *http.Cookie{
	spot := sm.StartSession(identifier)
	return newCookie(spot, identifier)
}

/*************************end session stuff****************************/

//First verifies that the session ending is correct (to prevent someone from messing
//around and ending other peoples sessions with some custom cookies), then ends the 
//session by: setting the spot active to false and adding that spot to the strung 
//together empty spots.
//
//more specific on the stringing: if sm.openSpot.chain == -1, then there is no spots in
//the string, so set it as sm.openSpot and as sm.lasedEndedSpot. other wise, set the 
//session at sm.lastEndedSpot to point to this newly-ended spot, and update 
//sm.lastEndedSpot
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

//parses the cookie passed, then calls sm.EndSession
func (sm *SessionManager) EndSessionCookie(c *http.Cookie) error {
	spot, identifier, err := sm.ParseCookie(c)
	if err != nil {
		return err	
	}
	err = sm.EndSession(spot, identifier)
	if err != nil {
		return err	
	}
	return nil
}

/***********************cookie parsing*******************************/

//parses , error checks, and returns accordingly. Cookie value should be of format:
//"6|965|brocrast21" (chain|index|identifier)
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


/********************resize stuff********************/
//called by sm.StartSession
//checks if resizing is needed. first if checks if its already being resized, then
//returns. Second if checks if the passed in spot (which is the newly started session)
//is at the sm.resizeAt variable, and it if was added to the last chain (because if 
//the spot.index is past the sm.resizeAt, but it was added to the first chain, and 
//there is 7 chains, then there is no need to resize yet). calls sm.resize
func (sm *SessionManager) checkResize(spot SpotMarker){
	if sm.beingResized{
		return	
	}
	if spot.index >= sm.resizeAt && spot.chain == sm.currentChains {
		sm.beingResized = true
		sm.resize()	
	}
}

//called by sm.checkResize(). incriments the sm.currentChains, creates a chain that is
//one size bigger, then copies over the contents, and sets sm.chains equal to newChains.
//resets sm.beingResized to false
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


/************************nextSpot**********************/

//next spot returns the next available spot for a session to be started. called by
//sm.StartSession(). 
//first checks if there is any strung together from ended sessions. If sm.openSpot.chain
// == -1, then that means there is no ended spots (done through consistency throughout 
//the program), so send back from the farthestPlaced. otherwise, send back from open spot
//, and update the strung together ended sessions accordingly
func (sm *SessionManager) nextSpot() SpotMarker {
	//lock
	//defer unlock
	if sm.openSpot.chain == -1 {
		if sm.farthestPlaced.index <= sm.chainSize - 2{
			sm.farthestPlaced.index++
		} else {
			sm.farthestPlaced.index = 0
			sm.farthestPlaced.chain++
		}
		return sm.farthestPlaced
	}
	spot := sm.openSpot
	if sm.lastEndedSpot == spot{
		sm.openSpot.chain = -1
	} else {
		sm.openSpot = sm.chains[spot.chain][spot.index].nextSpot
	}
	return spot
}


/**************************new cookie***********************/

//creates the value string by stringing things together with delimiter "|". will be
// "(chainSessionIsStoredIn)|(IndexInThatChain)|(identifier)", example: "4|314|pablo667"
//creates a cookiei with Name: "session" and Value: "6|7|hiker777"(or whatever)
func newCookie(spot SpotMarker, identifier string) *http.Cookie{
	val := strconv.Itoa(spot.chain) + "|" + strconv.Itoa(spot.index) + "|" + identifier
	c := &http.Cookie{Name: "session", Value: val}
	return c
}







