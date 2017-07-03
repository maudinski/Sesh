# Sesh

This is an easy to use, abstract session manager for Go.


There are four main functions:

	sm := sesh.NewSM() //returns SessionManager object with default settings
	sm.StartSession(w, uniqueString) //w is http.Response Writer, unique string is username or something unique (like whatever you're using to identify/store users in a database)
	sm.VerifySession(r) //r is *http.Request
	sm.EndSession(r) //r is *http.Request

and one alternative to NewSM()
	
	sm := sesh.NewCustomSM(1000) //will eventually have more functionality/custimization options


Under the hood:

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
