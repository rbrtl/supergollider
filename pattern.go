package supergollider

import (
	"fmt"
	"math/rand"
	"time"
)

type Pattern interface {
	Events(barNum int, t Tracker) map[Measure][]*Event
	NumBars() int
}

type slotPattern struct {
	relations []float64
	//factor     float64
	base   string
	events []interface{}
	at     Measure
}

// SlotPattern creates slots by scaling the given relations to the current Bar of the used tracker.
// The slots are Measure positions on which the events that are passed via SetEvents() are positioned.
// Each of events may be *Event or []*Event
// If len(relations) < len(events), the positions of the events will rotate through the slots.
// If len(events) < len(relations), the events will rotate through the slot positions.
// The scaling base can be customized by calling SetBase(). The default base is the current Bar.
// The pattern by default starts at position 0 inside the bar, but that could be changed by calling At()
func SlotPattern(relations ...float64) *slotPattern {
	return &slotPattern{relations: relations}
}

func (sp *slotPattern) SetEvents(events ...interface{}) *slotPattern {
	for _, x := range events {
		switch x.(type) {
		case *Event, []*Event:
		default:
			if x != nil {
				panic(fmt.Sprintf("type %T not allowed as event in SlotPatterns, just *Event and []*Event", x))
			}
		}
	}
	sp.events = events
	return sp
}

// SetEventsAt sets the events at a certain slot index
func (sp *slotPattern) SetEventsAt(idx int, evts ...*Event) *slotPattern {
	for len(sp.events) < idx+1 {
		sp.events = append(sp.events, nil)
	}
	sp.events[idx] = evts
	return sp
}

func (sp *slotPattern) MapEvents(m map[int]interface{}) *slotPattern {

	for idx, x := range m {
		switch e := x.(type) {
		case *Event:
			sp.SetEventsAt(idx, e)
		case []*Event:
			sp.SetEventsAt(idx, e...)
		}
	}

	return sp
}

func (sp *slotPattern) ChangeEventsAt(idx int, fn func(evts ...*Event)) *slotPattern {
	x := sp.events[idx%len(sp.events)]

	switch e := x.(type) {
	case *Event:
		fn(e)
	case []*Event:
		fn(e...)
	default:
	}

	return sp
}

func (sp *slotPattern) ChangeAllEvents(fn func(idx int, evts ...*Event)) *slotPattern {

	for i, x := range sp.events {
		switch e := x.(type) {
		case *Event:
			fn(i, e)
		case []*Event:
			fn(i, e...)
		default:
		}
	}

	return sp
}

func (sp *slotPattern) Clone() *slotPattern {
	return &slotPattern{
		relations: sp.relations,
		base:      sp.base,
		events:    sp.events,
		at:        sp.at,
	}
}

func (sp *slotPattern) SetBase(base string) *slotPattern {
	sp.base = base
	return sp
}

func (sp *slotPattern) At(m string) *slotPattern {
	sp.at = M(m)
	return sp
}

func (sp *slotPattern) Events(barNum int, t Tracker) map[Measure][]*Event {
	var base = t.CurrentBar()

	if sp.base != "" {
		base = M(sp.base)
	}

	ms := base.Scale(sp.relations...)
	res := map[Measure][]*Event{}
	last := -ms[0] + sp.at

	// fmt.Fprintf(Debug, "\n")

	if len(sp.events) >= len(ms) {
		// var last Measure
		for i, x := range sp.events {
			var evts []*Event

			if x != nil {
				switch ev := x.(type) {
				case *Event:
					evts = []*Event{ev.Clone()}
				case []*Event:
					evts = make([]*Event, len(ev))
					for j, e := range ev {
						evts[j] = e.Clone()
					}
				}

				last = ms[i%len(ms)] + last
				// fmt.Fprintf(Debug, "%s » ", last)
				res[last] = evts
			}
		}

		return res
	}

	// var last Measure
	for i, m := range ms {
		x := sp.events[i%len(sp.events)]

		if x != nil {
			var evts []*Event
			switch ev := x.(type) {
			case *Event:
				evts = []*Event{ev.Clone()}
			case []*Event:
				evts = make([]*Event, len(ev))
				for j, e := range ev {
					evts[j] = e.Clone()
				}
			}
			last = m + last
			// fmt.Fprintf(Debug, "%s | ", last)
			res[last] = evts
		}
	}

	return res
}

func (sp *slotPattern) NumBars() int {
	return 1
}

/*
type slotPattern struct {
	relations   []Measure
	events  []*Event
	i       int
	numBars int
	res     map[int]map[Measure][]*Event
}

func SlotPattern(relations []Measure, events ...*Event) Pattern {
	s := &slotPattern{relations, events}
	s.calculate()
	return s
}

func (s *slotPattern) calculate(t Tracker) {
	s.res = map[int]map[Measure][]*Event{}

	sumstr := "0"

	for _, sl := range s.relations {
		sumstr += " + " + sl.String()
	}

	numBars, _ := t.CurrentBar().Add(M(sumstr))
	numBars++
}

func (s *slotPattern) Events(barNum int, t Tracker) map[Measure][]*Event {
	return s.res[barNum]
}

func (s *slotPattern) NumBars() int {
	//return s.numBars
	return 1
}
*/

type PatternFunc func(barNum int, t Tracker) map[Measure][]*Event

func (tf PatternFunc) Events(barNum int, t Tracker) map[Measure][]*Event {
	return tf(barNum, t)
}

func (tf PatternFunc) NumBars() int {
	return 1
}

/*
type seqModTrafo struct {
	*seqPlay
	pos            Measure
	overrideParams Parameter
	params         Parameter
}

func (sm *seqModTrafo) Pattern(tr *Track) {
	tr.At(sm.pos, ChangeEvent(sm.seqPlay.v, Params(sm.params, sm.overrideParams)))
}

type seqPlay struct {
	seq        []Parameter
	initParams Parameter
	v          *Voice
	Pos        int
}

func (sp *seqPlay) Modify(pos string, params ...Parameter) Pattern {
	params_ := sp.seq[sp.Pos]
	if sp.Pos < len(sp.seq)-1 {
		sp.Pos++
	} else {
		sp.Pos = 0
	}
	return &seqModTrafo{seqPlay: sp, pos: M(pos), overrideParams: Params(params...), params: params_}
}

func (sp *seqPlay) PlayDur(pos, dur string, params ...Parameter) Pattern {
	params_ := sp.seq[sp.Pos]
	if sp.Pos < len(sp.seq)-1 {
		sp.Pos++
	} else {
		sp.Pos = 0
	}
	return &seqPlayTrafo{seqPlay: sp, pos: M(pos), dur: M(dur), overrideParams: Params(params...), params: params_}
}
*/

/*
func ParamSequence(v *Voice, initParams Parameter, paramSeq ...Parameter) *seqPlay {
	return &seqPlay{
		initParams: initParams,
		seq:        paramSeq,
		v:          v,
	}
}
*/

/*
type seqPlayTrafo struct {
	*seqPlay
	pos            Measure
	dur            Measure
	overrideParams Parameter
	params         Parameter
}
*/

/*
func (spt *seqPlayTrafo) Params() (p map[string]float64) {
	return Params(spt.seqPlay.initParams, spt.seqPlay.seq[spt.seqPlay.Pos], spt.overrideParams).Params()
}
*/

/*
func (spt *seqPlayTrafo) Pattern(tr *Track) {
	params := Params(spt.seqPlay.initParams, spt.params, spt.overrideParams)
	tr.At(spt.pos, OnEvent(spt.seqPlay.v, params))
	tr.At(spt.pos+spt.dur, OffEvent(spt.seqPlay.v))
*/
/*
	if spt.seqPlay.Pos < len(spt.seqPlay.seq)-1 {
		spt.seqPlay.Pos++
	} else {
		spt.seqPlay.Pos = 0
	}
*/
/*
}
*/

// type end struct{}

func (e End) Pattern(t Tracker) {
	t.At(M(string(e)), fin)
}

func (e End) Events(barNum int, t Tracker) map[Measure][]*Event {
	return map[Measure][]*Event{
		M(string(e)): []*Event{fin},
	}
}

func (e End) NumBars() int {
	return 1
}

type End string

type Start string

func (s Start) Events(barNum int, t Tracker) map[Measure][]*Event {
	return map[Measure][]*Event{
		M(string(s)): []*Event{start},
	}
}

func (s Start) NumBars() int {
	return 1
}

/*
type stopAll struct {
	pos    Measure
	Voices []*Voice
}

func StopAll(pos string, vs ...[]*Voice) *stopAll {
	s := &stopAll{pos: M(pos)}

	for _, v := range vs {
		s.Voices = append(s.Voices, v...)
	}

	return s
}

func (p *stopAll) Pattern(t Tracker) {
	for i := 0; i < len(p.Voices); i++ {
		t.At(p.pos, OffEvent(p.Voices[i]))
	}
}
*/

type setTempo struct {
	Tempo Tempo
	Pos   Measure
}

func SetTempo(at string, t Tempo) *setTempo {
	return &setTempo{t, M(at)}
}

// Tracker must be a *Track
/*
func (s *setTempo) Pattern(t Tracker) {
	t.(*Track).SetTempo(s.Pos, s.Tempo)
}
*/

func (s *setTempo) Events(barNum int, t Tracker) map[Measure][]*Event {
	return map[Measure][]*Event{
		s.Pos: []*Event{BPM(s.Tempo.BPM()).Event()},
	}
}

func (s *setTempo) NumBars() int {
	return 1
}

/*
 */

/*
type times struct {
	times int
	trafo Pattern
}

func (n *times) NumBars() int {
	return n.times
}

func (n *times) Events(barNum int, barMeasure Measure) map[Measure][]*Event {
	res := map[Measure][]*Event{}

	for i := 0; i < n.times; i++ {
		for pos, events := range n.trafo.Events(barNum, barMeasure) {
			finalPos := M(fmt.Sprintf("%d", i)) + pos
			res[finalPos] = append(res[finalPos], events...)
		}
	}
	return res
}
*/

/*
func (n *times) Pattern(t Tracker) {
	for i := 0; i < n.times; i++ {
		n.trafo.Pattern(t)
	}
}
*/

/*

func Times(num int, trafo Pattern) Pattern {
	return &times{times: num, trafo: trafo}
}
*/

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type randomPattern []Pattern

func (r randomPattern) Events(barNum int, t Tracker) map[Measure][]*Event {
	return r[rand.Intn(len(r))].Events(barNum, t)
}

func (r randomPattern) NumBars() int {
	return r[0].NumBars()
}

// each pattern must have the same NumBars
func RandomPattern(patterns ...Pattern) Pattern {
	if len(patterns) < 1 {
		panic("at least one pattern required")
	}
	numBars := patterns[0].NumBars()

	if len(patterns) == 1 {
		return randomPattern(patterns)
	}

	for i, p := range patterns[1:] {
		if p.NumBars() != numBars {
			panic(fmt.Sprintf("pattern %d does not have the same number of bars as pattern 0", i))
		}
	}

	return randomPattern(patterns)
}

type tempoSpan struct {
	start    float64
	current  float64
	step     float64
	modifier func(current, step float64) float64
	// t        *Track
}

type tempoSpanTrafo struct {
	*tempoSpan
	pos string
}

func (ts *tempoSpan) SetTempo(pos string) Pattern {
	return &tempoSpanTrafo{ts, pos}
}

func (ts *tempoSpan) Reset(pos string) Pattern {
	fn := func(barNum int, tr Tracker) map[Measure][]*Event {
		ts.current = ts.start
		return BPM(ts.current).Pattern(pos).Events(barNum, tr)
	}
	return PatternFunc(fn)
	/*
		return EventFuncPattern(pos, func(e *Event) {
			// fmt.Printf("resetting\n")
			ts.current = ts.start
		})
	*/
}

func (ts *tempoSpanTrafo) NumBars() int {
	return 1
}

// TODO: fix it! we need to have a TempoSpan event that is handled specially by the  Tracker - OR - a
// special Event type that has a method that receives a *Track and may act upon it
// that must be somehow integrated to the tracker methods
func (ts *tempoSpanTrafo) Events(barNum int, t Tracker) (res map[Measure][]*Event) {
	// var newtempo float64
	ts.current = ts.modifier(ts.current, ts.step)
	return BPM(ts.current).Pattern(ts.pos).Events(barNum, t)
	/*
		return map[Measure][]*Event{
			M(ts.pos): {SetBPM(pos, )},
		}
	*/
	/*
		if ts.current == -1 {
			//newtempo = ts.modifier(ts.t.TempoAt(M(ts.pos)).BPM(), ts.step)
			newtempo = ts.modifier(t.TempoAt(M(ts.pos)).BPM(), ts.step)
		} else {
			ts.current = ts.modifier(ts.current, ts.step)
			newtempo = ts.current
		}
		//rounded := RoundFloat(newtempo, 4)
		//ts.t.SetTempo(M(ts.pos), BPM(newtempo))
		t.SetTempo(M(ts.pos), BPM(newtempo))
	*/
	// return nil
}

/*
func TempoSpan(step float64, modifier func(current, step float64) float64) *tempoSpanTrafo {
	ts := &tempoSpan{current: 0, step: step, modifier: modifier}

}
*/
/*
// Tracker must be a *Track
func (ts *tempoSpanTrafo) Pattern(t Tracker) {
	var newtempo float64
	if ts.current == -1 {
		newtempo = ts.modifier(t.(*Track).TempoAt(M(ts.pos)).BPM(), ts.step)
	} else {
		ts.current = ts.modifier(ts.current, ts.step)
		newtempo = ts.current
	}
	//rounded := RoundFloat(newtempo, 4)
	t.(*Track).SetTempo(M(ts.pos), BPM(newtempo))
}
*/

func StepAdd(current, step float64) float64 {
	return current + step
}

func StepMultiply(current, step float64) float64 {
	return current * step
}

// for start = -1 takes the current tempo
func SeqTempo(start float64, step float64, modifier func(current, step float64) float64) *tempoSpan {
	return &tempoSpan{start: start, current: start, step: step, modifier: modifier}
}

type seqBool struct {
	seq   []bool
	pos   int
	trafo Pattern
}

func (s *seqBool) Events(barNum int, t Tracker) (res map[Measure][]*Event) {
	if s.trafo == nil {
		return nil
	}
	if s.seq[s.pos] {
		res = s.trafo.Events(barNum, t)
	}
	if s.pos < len(s.seq)-1 {
		s.pos++
	} else {
		s.pos = 0
	}
	return
}

func (s *seqBool) NumBars() int {
	if s.trafo == nil {
		return 1
	}
	return s.trafo.NumBars()
}

func SeqSwitch(trafo Pattern, seq ...bool) Pattern {
	return &seqBool{seq: seq, pos: 0, trafo: trafo}
}

type sequence []Pattern

func (s sequence) NumBars() int {
	num := 0

	for _, p := range s {
		num += p.NumBars()
	}
	return num
}

func (s sequence) Events(barNum int, t Tracker) (res map[Measure][]*Event) {
	num := 0

	for _, p := range s {
		next := num + p.NumBars()
		if barNum < next {
			return p.Events(barNum-num, t)
		}
		num = next
	}
	return
}

func SeqPatterns(seq ...Pattern) Pattern {
	return sequence(seq)
}

type compose []Pattern

func (c compose) Events(barNum int, t Tracker) map[Measure][]*Event {
	res := map[Measure][]*Event{}
	for _, pattern := range c {
		if pattern != nil {
			for pos, events := range pattern.Events(barNum, t) {
				res[pos] = append(res[pos], events...)
			}
		}
	}
	return res
}

func (c compose) NumBars() int {
	max := 1
	for _, pattern := range c {
		if pattern != nil {
			if pattern.NumBars() > max {
				max = pattern.NumBars()
			}
		}
	}
	return max
}

func MixPatterns(trafos ...Pattern) Pattern {
	return compose(trafos)
}

type linearDistribute struct {
	from, to float64
	steps    int
	dur      Measure
	key      string
	// from, to float64, steps int, dur Measure
}

type linearTempoLine struct {
	*linearDistribute
}

func LinearTempoChange(from, to float64, n int, dur Measure) *linearTempoLine {
	return &linearTempoLine{LinearDistribution("bpm", from, to, n, dur)}
}

func (l *linearTempoLine) ModifyTempo(position string) Pattern {
	return l.modifyTempo(position)
}

// LinearDistribution creates a transformer that modifies the given parameter param
// from the value from to the value to in n steps in linear growth for a total duration dur
func LinearDistribution(param string, from, to float64, n int, dur Measure) *linearDistribute {
	return &linearDistribute{from, to, n, dur, param}
}

func (l *linearDistribute) modifyTempo(position string) Pattern {
	p := []Pattern{}
	width, diff := LinearDistributedValues(l.from, l.to, l.steps, l.dur)
	val := l.from
	pos := M(position)
	for i := 0; i < l.steps; i++ {
		// println(pos.String())
		p = append(p, &setBpm{pos, BPM(val)})
		// tr.At(pos, ChangeEvent(ld.v, ParamsMap(map[string]float64{ld.linearDistribute.key: val})))
		pos += width
		val += diff
	}

	return MixPatterns(p...)
}

func (l *linearDistribute) ModifyVoice(position string, v *Voice) Pattern {
	//return &linearDistributeTrafo{l, v, M(position)}
	p := []Pattern{}
	width, diff := LinearDistributedValues(l.from, l.to, l.steps, l.dur)
	// tr.At(ld.pos, Change(ld.v, ))
	pos := M(position)
	val := l.from
	for i := 0; i < l.steps; i++ {
		// println(pos.String())
		p = append(p, &mod{pos, v, ParamsMap(map[string]float64{l.key: val})})
		// tr.At(pos, ChangeEvent(ld.v, ParamsMap(map[string]float64{ld.linearDistribute.key: val})))
		pos += width
		val += diff
	}

	return MixPatterns(p...)
}

type linearDistributeTrafo struct {
	*linearDistribute
	v   *Voice
	pos Measure
}

/*
func (ld *linearDistributeTrafo) Pattern(tr *Track) {
	width, diff := LinearDistributedValues(ld.linearDistribute.from, ld.linearDistribute.to, ld.linearDistribute.steps, ld.linearDistribute.dur)
	// tr.At(ld.pos, Change(ld.v, ))
	pos := ld.pos
	val := ld.linearDistribute.from
	for i := 0; i < ld.linearDistribute.steps; i++ {
		tr.At(pos, ChangeEvent(ld.v, ParamsMap(map[string]float64{ld.linearDistribute.key: val})))
		pos += width
		val += diff
	}
}
*/

// ---------------------------------------

// func ExponentialDistributedValues(from, to float64, steps int, dur Measure) (width Measure, diffs []float64) {

type expDistribute struct {
	from, to float64
	steps    int
	dur      Measure
	key      string
	// from, to float64, steps int, dur Measure
}

// ExponentialDistribution creates a transformer that modifies the given parameter param
// from the value from to the value to in n steps in exponential growth for a total duration dur
func ExponentialDistribution(param string, from, to float64, n int, dur Measure) *expDistribute {
	return &expDistribute{from, to, n, dur, param}
}

func (l *expDistribute) ModifyVoice(position string, v *Voice) Pattern {
	p := []Pattern{}
	width, diffs := ExponentialDistributedValues(l.from, l.to, l.steps, l.dur)
	// tr.At(ld.pos, Change(ld.v, ))
	pos := M(position)
	for i := 0; i < l.steps; i++ {
		p = append(p, &mod{pos, v, ParamsMap(map[string]float64{l.key: diffs[i]})})
		//tr.At(pos, ChangeEvent(ld.v, ParamsMap(map[string]float64{ld.expDistribute.key: diffs[i]})))
		pos += width
		//val += diff
	}
	//return &expDistributeTrafo{l, v, M(position)}
	return MixPatterns(p...)
}

func (l *expDistribute) modifyTempo(position string) Pattern {
	p := []Pattern{}
	width, diffs := ExponentialDistributedValues(l.from, l.to, l.steps, l.dur)
	pos := M(position)
	for i := 0; i < l.steps; i++ {
		// println(pos.String())
		p = append(p, &setBpm{pos, BPM(diffs[i])})
		// tr.At(pos, ChangeEvent(ld.v, ParamsMap(map[string]float64{ld.linearDistribute.key: val})))
		pos += width
		// val += diff
	}

	return MixPatterns(p...)
}

type expTempoLine struct {
	*expDistribute
}

func ExponentialTempoChange(from, to float64, n int, dur Measure) *expTempoLine {
	return &expTempoLine{ExponentialDistribution("bpm", from, to, n, dur)}
}

func (l *expTempoLine) ModifyTempo(position string) Pattern {
	return l.modifyTempo(position)
}

type expDistributeTrafo struct {
	*expDistribute
	v   *Voice
	pos Measure
}

/*
func (ld *expDistributeTrafo) Pattern(tr *Track) {
	width, diffs := ExponentialDistributedValues(ld.expDistribute.from, ld.expDistribute.to, ld.expDistribute.steps, ld.expDistribute.dur)
	// tr.At(ld.pos, Change(ld.v, ))
	pos := ld.pos
	for i := 0; i < ld.expDistribute.steps; i++ {
		tr.At(pos, ChangeEvent(ld.v, ParamsMap(map[string]float64{ld.expDistribute.key: diffs[i]})))
		pos += width
		//val += diff
	}
}
*/
