// Copyright 2017 Nicholas Ng <nickng@nickng.io>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package aut provides conversion functions for the AUT format
// (AUTomation, or ALDEBARAN format) labelled transition systems
// to and from the sesstype language.
package aut

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/nickng/aut"
	"go.nickng.io/sesstype"
	"go.nickng.io/sesstype/local"
)

var (
	ErrInvalidLocalType = errors.New("toAut: cannot convert unknown local type")
	ErrInvalidLabel     = errors.New("fromAut: invalid transition label")
	ErrInconsistent     = errors.New("fromAut: transitions inconsistent")
)

type conv struct {
	fsm       *aut.Aut
	scope     map[string]aut.State
	nstates   int
	nextstate int
}

func newConv() *conv {
	return &conv{
		fsm:   new(aut.Aut),
		scope: make(map[string]aut.State),
	}
}

func (c *conv) newState() aut.State {
	c.nextstate++
	c.nstates++
	return aut.State(c.nextstate - 1)
}

func (c *conv) toAut(from aut.State, t local.Type) (aut.State, error) {
	switch t := t.(type) {
	case *local.Branch:
		for m, l := range t.Locals {
			st := c.newState()
			next, err := c.toAut(st, l)
			if err != nil {
				return 0, err
			}
			c.fsm.AddTransition(from, fmt.Sprintf("%s ? %s", t.From.Name(), m.String()), next)
		}
		return from, nil
	case *local.Select:
		for m, l := range t.Locals {
			st := c.newState()
			next, err := c.toAut(st, l)
			if err != nil {
				return 0, err
			}
			c.fsm.AddTransition(from, fmt.Sprintf("%s ! %s", t.To.Name(), m.String()), next)
		}
		return from, nil
	case *local.Recur:
		c.scope[t.T] = from
		return c.toAut(from, t.L)
	case local.TypeVar:
		st, ok := c.scope[t.T]
		if !ok {
			log.Fatalf("TypeVar %s does not refer to an existing scope", t.T)
		}
		c.nstates--
		return st, nil // from state dropped
	case local.End:
		return from, nil
	default:
		return 0, ErrInvalidLocalType
	}
}

// FromLocal takes as input a local Type and return a valid Aut state machine.
func FromLocal(l local.Type) *aut.Aut {
	c := newConv()
	initState := c.newState()
	c.toAut(initState, l)
	c.fsm.SetDes(int(initState), len(c.fsm.Transitions), c.nstates)
	return c.fsm
}

var (
	recv = regexp.MustCompile(`(\w+)\s*\?\s*(\w*)\s*\((.*)\)$`)
	send = regexp.MustCompile(`(\w+)\s*!\s*(\w*)\s*\((.*)\)`)
)

type dir int

const (
	Send dir = iota + 1
	Recv
)

func parseLabel(label string) (sesstype.Role, dir, sesstype.Message, error) {
	if recv.MatchString(label) {
		match := recv.FindAllStringSubmatch(label, -1)
		role := sesstype.NewRole(match[0][1])
		mesg := sesstype.Message{Label: match[0][2]}
		// Try parse as local type
		if s, err := local.Parse(strings.NewReader(match[0][3])); err != nil {
			mesg.Payload = sesstype.BaseType{Type: match[0][3]}
		} else {
			mesg.Payload = s
		}
		return role, Recv, mesg, nil
	} else if send.MatchString(label) {
		match := send.FindAllStringSubmatch(label, -1)
		role := sesstype.NewRole(match[0][1])
		mesg := sesstype.Message{Label: match[0][2], Payload: sesstype.BaseType{Type: match[0][3]}}
		// Try parse as local type
		if s, err := local.Parse(strings.NewReader(match[0][3])); err != nil {
			mesg.Payload = sesstype.BaseType{Type: match[0][3]}
		} else {
			mesg.Payload = s
		}
		return role, Send, mesg, nil
	}
	return sesstype.Role{}, 0, sesstype.Message{}, ErrInvalidLabel
}

// transition is a temporarily struct to help build the Type.
type transition struct {
	role  sesstype.Role
	dir   dir
	edges []struct {
		mesg sesstype.Message
		next aut.State
	}
}

// addEdge adds a message and next state to an existing transition.
func (tr *transition) addEdge(mesg sesstype.Message, next aut.State) {
	tr.edges = append(tr.edges, struct {
		mesg sesstype.Message
		next aut.State
	}{
		mesg: mesg,
		next: next,
	})
}

// ToLocal takes as input an Aut state machine and return a valid local Type.
func ToLocal(a *aut.Aut) (local.Type, error) {
	trans, inEdges := make(map[aut.State]*transition), make(map[aut.State]int)
	for _, tr := range a.Transitions {
		r, d, m, err := parseLabel(tr.Label)
		if err != nil {
			return nil, err
		}
		if _, exists := trans[tr.From]; !exists {
			trans[tr.From] = &transition{role: r, dir: d}
		}
		trans[tr.From].addEdge(m, tr.To)
		inEdges[tr.To]++
	}
	l := stateToType(a.Init, trans, inEdges, make(map[aut.State]string))
	return l, nil
}

// stateToType converts a state into a local Type.
func stateToType(st aut.State, trans map[aut.State]*transition, inEdges map[aut.State]int, labels map[aut.State]string) local.Type {
	if _, exists := trans[st]; !exists {
		return local.NewEnd()
	}
	var isrecur bool
	if inEdges[st] > 0 {
		if _, exists := labels[st]; exists {
			// Using a previously created recur label, i.e. this is a typevar.
			return local.NewTypeVar(labels[st])
		} else {
			// Defining a recur label, i.e. this is a μT.L.
			labels[st] = fmt.Sprintf("L%d", int(st))
			isrecur = true
		}
	}
	var typ local.Type
	m := make(map[sesstype.Message]local.Type)
	for i := range trans[st].edges {
		m[trans[st].edges[i].mesg] = stateToType(trans[st].edges[i].next, trans, inEdges, labels)
	}
	switch trans[st].dir {
	case Send:
		typ = local.NewSelect(trans[st].role, m)
	case Recv:
		typ = local.NewBranch(trans[st].role, m)
	}
	if isrecur {
		// Instead of returning normal statement, wrap with μT.L
		return local.NewRecur(labels[st], typ)
	}
	return typ
}
