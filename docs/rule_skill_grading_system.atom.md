---
id: rule_skill_grading_system
status: STABLE
human_name: Skill Grading System Rule
type: RULE
layer: ARCHITECTURE
version: 2.0
priority: 5
tags: [skills, grading, progression]
parents: []
dependents:
  - [[upsilonapi:api_skill_grading_computation]]
---

# New Atom

## INTENT
To establish the skill grading system where skill grades (I-V) are determined by Total Positive Skill Weight (SW), providing a clear progression path and power scaling for skills.

## THE RULE / LOGIC
Skill grades are determined by Total Positive Skill Weight (the sum of all benefits, ignoring costs):

**Grade Thresholds:**
- **Grade I:** 0 - 150 Positive SW (Basic skills, simple effects)
- **Grade II:** 151 - 300 Positive SW (Intermediate skills, moderate power)
- **Grade III:** 301 - 500 Positive SW (Advanced skills, significant effects)
- **Grade IV:** 501 - 750 Positive SW (Expert skills, powerful combinations)
- **Grade V:** 750+ Positive SW (Ultimate skills, game-changing power)

**Credit Cost Calculation:**
- Credit Cost = Total Positive SW × 2
- Example: Grade I skill (100 SW) = 200 credits
- Example: Grade V skill (800 SW) = 1600 credits

**Progression Availability:**
- Character Creation: Grade I-II skills offered
- Every 10 Levels: Higher grade skills become available
- Level 1-9: Grade I-II skills
- Level 10-19: Grade II-III skills  
- Level 20-29: Grade III-IV skills
- Level 30+: Grade IV-V skills

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[rule_skill_grading_system]]`
- **Test Names:** `TestSkillGradeCalculation`, `TestSkillCreditCostCalculation`

## EXPECTATION
