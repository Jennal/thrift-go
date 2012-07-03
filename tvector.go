/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package thrift

/**
 * Helper class that encapsulates list metadata.
 *
 */
type TVector interface {
  TContainer
  ElemType() TType
  At(i int) interface{}
  Set(i int, data interface{})
  Push(data interface{})
  Pop() interface{}
  Swap(i, j int)
  Insert(i int, data interface{})
  Delete(i int)
  Less(i, j int) bool
  Iter() <-chan interface{}

  Last() interface{}
}

type tVector struct {
  elemType TType
  l        []interface{}
}

func NewTVector(t TType, s int) TVector {
  var v []interface{} = make([]interface{}, s)
  return &tVector{elemType: t, l: v}
}

func NewTVectorDefault() TVector {
  var v []interface{}
  return &tVector{elemType: TType(STOP), l: v}
}

func (p *tVector) ElemType() TType {
  return p.elemType
}

func (p *tVector) Len() int {
  return len(p.l)
}

func (p *tVector) At(i int) interface{} {
  return p.l[i]
}

func (p *tVector) Set(i int, data interface{}) {
  if p.elemType.IsEmptyType() {
    p.elemType = TypeFromValue(data)
  }
  if data, ok := p.elemType.CoerceData(data); ok {
    p.l[i] = data
  }
}

func (p *tVector) Push(data interface{}) {
  if p.elemType.IsEmptyType() {
    p.elemType = TypeFromValue(data)
  }
  data, ok := p.elemType.CoerceData(data)
  if ok {
    p.l = append(p.l, data)
  }
}

func (p *tVector) Pop() interface{} {
  var v interface{}
  v, p.l = p.l[len(p.l)-1], p.l[:len(p.l)-1]
  return v
}

func (p *tVector) Swap(i, j int) {
  p.l[i], p.l[j] = p.l[j], p.l[i]
}

func (p *tVector) Insert(i int, data interface{}) {
  left  := p.l[:i]
  right := p.l[i:]
  p.l = append(left, data)
  p.l = append(p.l, right)
}

func (p *tVector) Delete(i int) {
  p.l = append(p.l[:i], p.l[i+1:])
}

func (p *tVector) Contains(data interface{}) bool {
  return p.indexOf(data) >= 0
}

func (p *tVector) Less(i, j int) bool {
  cv, _ := p.elemType.Compare(p.l[i], p.l[j])
  return cv < 0
}

func (p *tVector) Iter() <-chan interface{} {
  c := make(chan interface{})
  go p.iterate(c)
  return c
}

func (p *tVector) iterate(c chan<- interface{}) {
  for _, elem := range p.l {
    c <- elem
  }
  close(c)
}

func (p *tVector) indexOf(data interface{}) int {
  if data == nil {
    size := p.Len()
    for i := 0; i < size; i++ {
      if p.At(i) == nil {
        return i
      }
    }
    return -1
  }
  data, ok := p.elemType.CoerceData(data)
  if data == nil || !ok {
    return -1
  }
  size := p.Len()
  if p.elemType.IsBaseType() || p.elemType.IsEnum() {
    for i := 0; i < size; i++ {
      if data == p.At(i) {
        return i
      }
    }
    return -1
  }
  if cmp, ok := data.(EqualsOtherInterface); ok {
    for i := 0; i < size; i++ {
      if cmp.Equals(p.At(i)) {
        return i
      }
    }
    return -1
  }
  switch p.elemType {
  case MAP:
    if cmp, ok := data.(EqualsMap); ok {
      for i := 0; i < size; i++ {
        v := p.At(i)
        if v == nil {
          continue
        }
        if cmp.Equals(v.(TMap)) {
          return i
        }
      }
      return -1
    }
  case SET:
    if cmp, ok := data.(EqualsSet); ok {
      for i := 0; i < size; i++ {
        v := p.At(i)
        if v == nil {
          continue
        }
        if cmp.Equals(v.(TSet)) {
          return i
        }
      }
      return -1
    }
  case LIST:
    if cmp, ok := data.(EqualsList); ok {
      for i := 0; i < size; i++ {
        v := p.At(i)
        if v == nil {
          continue
        }
        if cmp.Equals(v.(TList)) {
          return i
        }
      }
      return -1
    }
  case STRUCT:
    if cmp, ok := data.(EqualsStruct); ok {
      for i := 0; i < size; i++ {
        v := p.At(i)
        if v == nil {
          continue
        }
        if cmp.Equals(v.(TStruct)) {
          return i
        }
      }
      return -1
    }
  }
  return -1
}

func (p *tVector) Equals(other interface{}) bool {
  c, cok := p.CompareTo(other)
  return cok && c == 0
}

func (p *tVector) CompareTo(other interface{}) (int, bool) {
  return TType(LIST).Compare(p, other)
}

func (p *tVector) Last() interface{} {
  if p.Len() <= 0 {
    return nil
  }

  return p.l[p.Len()-1]
}