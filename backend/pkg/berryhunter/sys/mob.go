package sys

import (
	"log"
	"math/rand"

	"github.com/EngoEngine/ecs"

	"github.com/trichner/berryhunter/pkg/berryhunter/gen"
	"github.com/trichner/berryhunter/pkg/berryhunter/items/mobs"
	"github.com/trichner/berryhunter/pkg/berryhunter/model"
	"github.com/trichner/berryhunter/pkg/berryhunter/model/mob"
	"github.com/trichner/berryhunter/pkg/berryhunter/phy"
	"github.com/trichner/berryhunter/pkg/berryhunter/wrand"
)

type MobSystem struct {
	mobs []model.MobEntity
	game model.Game
	rnd  *rand.Rand
}

func NewMobSystem(g model.Game, seed int64) *MobSystem {
	rnd := rand.New(rand.NewSource(seed))
	return &MobSystem{game: g, rnd: rnd}
}

func (n *MobSystem) Priority() int {
	return 20
}

func (n *MobSystem) New(w *ecs.World) {
	log.Println("MobSystem nominal")
}

func (n *MobSystem) AddEntity(e model.MobEntity) {
	n.mobs = append(n.mobs, e)
}

func (n *MobSystem) Update(dt float32) {
	for _, mob := range n.mobs {
		alive := mob.Update(dt)
		if !alive {
			n.game.RemoveEntity(mob.Basic())
			n.respawnMob(mob.MobDefinition())
		}
	}
}

func (n *MobSystem) respawnMob(d *mobs.MobDefinition) {
	m := mob.NewMob(d, false, n.game.Radius())

	if d.Generator.RespawnBehavior == mobs.RespawnBehaviorProcreation {
		randomMob := n.randomMob(d.ID)
		if randomMob != nil {
			m.SetPosition(n.findNearbySpawnPosition(randomMob.Position(), m, n.mobs))
			m.SetAngle(randomMob.Angle())
		} else {
			m.SetPosition(n.findMobSpawnPosition(m, n.mobs))
			m.SetAngle(0)
		}
	} else {
		m.SetPosition(n.findMobSpawnPosition(m, n.mobs))
		m.SetAngle(0)
	}

	n.game.AddEntity(m)
}

func (n *MobSystem) findMobSpawnPosition(spawned model.MobEntity, existing []model.MobEntity) phy.Vec2f {
	const maxTries = 64
	const minDistancePadding = 0.05

	best := gen.NewRandomPos(n.game.Radius())
	bestPenalty := float32(1e9)

	for i := 0; i < maxTries; i++ {
		candidate := gen.NewRandomPos(n.game.Radius())
		penalty := float32(0)
		isOverlapping := false
		for _, other := range existing {
			needed := spawned.Radius() + other.Radius() + minDistancePadding
			d2 := candidate.Sub(other.Position()).AbsSq()
			if d2 < needed*needed {
				isOverlapping = true
				penalty += needed*needed - d2
			}
		}
		if !isOverlapping {
			return candidate
		}
		if penalty < bestPenalty {
			best = candidate
			bestPenalty = penalty
		}
	}
	return best
}

func (n *MobSystem) findNearbySpawnPosition(center phy.Vec2f, spawned model.MobEntity, existing []model.MobEntity) phy.Vec2f {
	const maxTries = 64
	const minDistancePadding = 0.05
	const nearMin = 0.35
	const nearMax = 1.2

	best := center
	bestPenalty := float32(1e9)
	worldRadius2 := n.game.Radius() * n.game.Radius()

	for i := 0; i < maxTries; i++ {
		angle := n.rnd.Float32() * 2 * 3.1415927
		dist := nearMin + n.rnd.Float32()*(nearMax-nearMin)
		offset := phy.NewPolarVec2f(dist, angle)
		candidate := center.Add(offset)
		if candidate.AbsSq() > worldRadius2 {
			continue
		}

		penalty := float32(0)
		isOverlapping := false
		for _, other := range existing {
			needed := spawned.Radius() + other.Radius() + minDistancePadding
			d2 := candidate.Sub(other.Position()).AbsSq()
			if d2 < needed*needed {
				isOverlapping = true
				penalty += needed*needed - d2
			}
		}
		if !isOverlapping {
			return candidate
		}
		if penalty < bestPenalty {
			best = candidate
			bestPenalty = penalty
		}
	}

	if bestPenalty < float32(1e9) {
		return best
	}
	return n.findMobSpawnPosition(spawned, existing)
}

func (n *MobSystem) randomMob(id mobs.MobID) model.MobEntity {
	choices := []wrand.Choice{}
	for _, m := range n.mobs {
		if m.MobID() == id {
			choices = append(choices, wrand.Choice{Weight: 1, Choice: m})
		}
	}
	wc := wrand.NewWeightedChoice(choices)
	selected := wc.Choose(n.rnd)
	if selected == nil {
		return nil
	}
	return selected.(model.MobEntity)
}

func (n *MobSystem) Remove(b ecs.BasicEntity) {
	var delete int = -1
	for index, entity := range n.mobs {
		if entity.Basic().ID() == b.ID() {
			delete = index
			break
		}
	}
	if delete >= 0 {
		// e := p.players[delete]
		n.mobs = append(n.mobs[:delete], n.mobs[delete+1:]...)
	}
}
