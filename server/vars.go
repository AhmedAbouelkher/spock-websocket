package main

func IntVar(v int) *int             { return &v }
func StringVar(v string) *string    { return &v }
func BoolVar(v bool) *bool          { return &v }
func Float64Var(v float64) *float64 { return &v }
