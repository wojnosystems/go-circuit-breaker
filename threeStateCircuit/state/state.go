//go:generate go-enum --file=$GOFILE -noprefix

package state

// State allowed by the twoStateBreaker
/* ENUM(
Closed,
Open,
HalfOpen
)
*/
type State uint8
