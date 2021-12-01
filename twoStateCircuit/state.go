//go:generate go-enum --file=$GOFILE -noprefix --prefix State

package twoStateCircuit

// State allowed by the twoStateBreaker
/* ENUM(
Closed,
Open
)
*/
type State uint8
