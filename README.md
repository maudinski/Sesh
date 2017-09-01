# Sesh
vcfscfds
Session manager for Go.


Four main functions:

	sm := sesh.NewSM()
	sm.StartSession(w, identifier)
	sm.VerifySession(r) 
	sm.EndSession(r) 

and one alternative to NewSM()
	
	sm := sesh.NewCustomSM(1000)

----

# Function prototypes

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

