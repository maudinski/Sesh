# Sesh

This is an easy to use, abstract session manager for Go.


There are four main functions:

	sm := sesh.NewSM()
	sm.StartSession(w, identifier)
	sm.VerifySession(r) 
	sm.EndSession(r) 

and one alternative to NewSM()
	
	sm := sesh.NewCustomSM(1000)

----

# FunctionPrototypes and explanations:

	func NewSM() *SessionManager
	
returns a SessionManager object with default settings. Recommended

	func (sm *SessionManager) StartSession(w http.ResponseWriter, identifier string)
	
Starts a session for the enetered identifier by creating a cookie with some information
as the (inlcuding the identifier) and setting it to the http.ResponseWriter.  identifier 
is username or something unique (like whatever is used to identify/store users in a 
database)

	func (sm *SessionManager)VerifySession(r *http.Request) error

Verifies that a session exist/is valid. Parses the cookie value in the request. Returns
an error if the session doesnt exist or is invalid

	func (sm *SessionManager)EndSession(r *http.Request) error

Ends a session. Parses the cookie value stored in the request. Returns an error if the
session wasnt valid to begin with

	func NewCustomSM(chainSize int) *SessionManager 

Takes the initial size of the chains (that store the sessions). will eventually have 
more functionality/custimization options

----

# Under the hood:

The sessions are stored as a 2 dimensional slice. Its more specifically a slice
of chains, and each chain holds (by default) 1000 session structs. The position
of the session and the identifier(passed in to StartSession) are stored in a cookie
and set to the http.ResponseWriter.


Resizing is done when 1/2 of the last chain is full, and adds one more chain to the 
chains slice, then copies over the contents. The amount of chains added on resize and
when to resize will soon be customizable with NewCustomSM.


Ended sessions are strung together in a similar fashion to a linked list (but not 
using allocated memory). 

Starting a session requests a spot in the data structure from nextSpot() (not exported).
nextSpot first grabs from the Ended sessions list. If that list is empty, then it returns
from the top of the data structure. 
