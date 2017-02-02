// This part of the package ensures that authenticated users do not access parts of
// the site that they do not have roles for
package firewall

import (
	"net/http"
)

// Our type for specifying role flags
// These should be defined as such:
//
// const(
//		ROLE_1 Role = 1 << iota
//		ROLE_2
//		...
//		ROLE_N
// )
//
// Each route can have multiple rows assigned to it. By assigning roles to bit flags,
// it allows us to check all rows much quicker
type Role int

type Rollwall struct {
	route_rules map[string]Role
}

// Implements our Mechanism Inteface for Davinci
func (self *Rollwall) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

// Checks our route rules vs what our session currently tells us about the user
func (self *Rollwall) Check(r *http.Request) {

}
