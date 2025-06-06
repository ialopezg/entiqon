// Code generated by ENTIQON.
// File: internal/core/contracts/kindable.go
// Description: Defines the Kindable interface, representing any component that can expose and assign an internal Kind classification.
// Since: v1.6.0

package contract

// Kindable represents any object that provides type‐safe classification via a Kind value.
// Implementers should expose both GetKind() and SetKind(Kind) so that their classification
// can be inspected and modified by higher‐level logic.
//
// Typical implementers include token types like Column, Table, Condition, etc.
//
// # Example
//
//	var k Kindable = token.NewColumn("id")
//	k.SetKind(ColumnKind)
//	fmt.Println(k.GetKind()) // → ColumnKind
type Kindable interface {
	// GetKind returns the current Kind value for this object.
	// If no kind has been set or the receiver is nil, implementations should return UnknownKind.
	//
	// # Example
	//
	//     b := token.NewBaseToken("")
	//     fmt.Println(b.GetKind()) // → UnknownKind
	//     b.SetKind(TableKind)
	//     fmt.Println(b.GetKind()) // → TableKind
	GetKind() Kind

	// SetKind assigns a Kind value to this object.
	// This should be a no‐op if the receiver is nil.
	//
	// # Example
	//
	//     b := token.NewBaseToken("")
	//     b.SetKind(ColumnKind)
	//     fmt.Println(b.GetKind()) // → ColumnKind
	SetKind(k Kind)
}
