//go:generate go-enum --file=$GOFILE -noprefix

package state

// State allowed by the twoStateBreaker
/* ENUM(
Closed,
Open
)
*/
type State uint8
