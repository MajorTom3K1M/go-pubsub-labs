package main

func (u user) march(p piece, publishCh chan<- move) {
	publishCh <- move{userName: u.name, piece: p}
}

type user struct {
	name   string
	pieces []piece
}

type move struct {
	userName string
	piece    piece
}

type piece struct {
	location string
	name     string
}

func (u user) doBattles(subCh <-chan move) []piece {
	fights := []piece{}
	for mv := range subCh {
		if u.name == mv.userName {
			continue
		}
		for _, piece := range u.pieces {
			if piece.location == mv.piece.location {
				fights = append(fights, piece)
			}
		}
	}

	return fights
}

func distributeBattles(publishCh <-chan move, subChans []chan move) {
	for mv := range publishCh {
		for _, subCh := range subChans {
			subCh <- mv
		}
	}
}
