package player

import (
	"log"

	"github.com/trichner/berryhunter/pkg/berryhunter/model"
	"github.com/trichner/berryhunter/pkg/berryhunter/model/vitals"
)

func (p *player) Update(dt float32) {
	auraCollisions := p.damageAura.Collisions()
	for c := range auraCollisions {
		usr := c.Shape().UserData
		if usr == nil {
			log.Printf("Missing UserData!")
			continue
		}

		r, ok := usr.(model.Interacter)
		if !ok {
			log.Printf("Non conformant UserData: %T", usr)
			continue
		}
		r.PlayerTouches(p, p.config.DamageAuraDamageFraction)
	}

	// update time based tings

	// action
	if p.ongoingAction != nil {
		a := p.ongoingAction
		a.Update(dt)
		if a.TicksRemaining() < 0 {
			p.ongoingAction = nil
		}
	}

	if !p.isGod {
		p.updateVitalSigns(dt)
	}
}

func (p *player) updateVitalSigns(dt float32) {
	vitalSigns := p.VitalSigns()
	c := p.config

	// Hunger and cold are disabled: keep both vital signs maxed to prevent
	// starving/freezing damage and related status effects.
	vitalSigns.Satiety = vitals.Max
	vitalSigns.BodyTemperature = vitals.Max

	// Keep normal health regeneration behavior.
	if vitalSigns.Health != vitals.Max {
		healthFraction := c.HealthGainTick
		p.addHealthFraction(healthFraction)
		p.statusEffects.Add(model.StatusEffectRegenerating)
	}
}

func (p *player) addHealthFraction(fraction float32) {
	h := p.VitalSigns().Health
	h = h.AddFraction(fraction)
	p.VitalSigns().Health = h
}
