package wasmtime

type slab struct {
  list []int
  next int
}

func (s *slab) allocate() int {
  if s.next == len(s.list) {
    s.list = append(s.list, s.next)
  }
  ret := s.next
  s.next = s.list[ret]
  return ret
}

func (s *slab) deallocate(slot int) {
  s.list[slot] = s.next
  s.next = slot
}
