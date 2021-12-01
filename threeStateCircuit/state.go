//go:generate go-enum --file=$GOFILE -noprefix --prefix State

package threeStateCircuit

// State allowed by the twoStateBreaker
/* ENUM(
Closed,
Open,
HalfOpen
)
*/
type State uint8
