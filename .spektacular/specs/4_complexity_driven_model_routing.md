# Feature: Complexity-Driven Model Routing

## Overview
Automatically analyze specification complexity and route to appropriate model tier (Haiku/Sonnet/Opus) with fast/detailed mode selection for optimal speed and cost efficiency. This feature will make the planning phase smarter by using cheaper, faster models for simple specs while reserving expensive, powerful models for complex specifications.

## Requirements
- [ ] Implement spec complexity analyzer that scores specifications from 0.0 (simple) to 1.0 (complex)
- [ ] Integrate complexity analysis into the plan command before agent execution
- [ ] Route to appropriate model tier based on existing config.yaml complexity thresholds
- [ ] Support fast mode (minimal research) for simple specs and detailed mode for complex specs
- [ ] Provide upgrade path from fast to detailed mode if user needs more depth
- [ ] Log complexity scores and model routing decisions for analysis and tuning

## Constraints
- Must use existing Pydantic configuration models (SpektacularConfig.complexity, SpektacularConfig.models)
- Must work with existing planner agent without modifying agent specification
- Must maintain backward compatibility with manual model selection
- Cannot break existing spektacular plan command interface
- Complexity analysis must complete in <1 second for reasonable spec sizes

## Acceptance Criteria
- [ ] Simple specs (CSS changes, config updates) route to Haiku with fast mode automatically
- [ ] Medium specs (new features, integrations) route to Sonnet with detailed mode
- [ ] Complex specs (architectural changes, multi-phase features) route to Opus with detailed mode
- [ ] spektacular plan --model override still works for manual model selection
- [ ] Complexity scores are logged and can be reviewed for tuning thresholds
- [ ] Fast mode plans can be upgraded to detailed mode preserving user edits
- [ ] Performance: 80%+ cost reduction on simple specs, <10% speed degradation on complex specs

## Technical Approach

### Complexity Analysis Algorithm
Analyze spec content for complexity indicators:
- **Length factors**: Word count, line count, section count
- **Technical complexity**: Technical terminology density, API mentions, database references
- **Scope indicators**: Number of requirements, acceptance criteria count
- **Integration complexity**: External service mentions, multi-system interactions
- **Uncertainty markers**: "TBD", "unclear", question marks, "investigate" keywords

### Implementation Location
- Add `complexity.py` module with analyzer functions
- Extend `plan.py` to call analyzer before agent execution
- Update CLI to support `--complexity-override` flag for manual control
- Add complexity score to plan.md metadata for reference

### Model Routing Logic
```python
if complexity < 0.3:
    model = config.models.tiers.simple  # Haiku
    mode = "fast"
elif complexity < 0.7:
    model = config.models.tiers.medium  # Sonnet  
    mode = "detailed"
else:
    model = config.models.tiers.complex  # Opus
    mode = "detailed"
```

### Integration Points
- Extend planner agent to support fast/detailed mode instructions
- Use existing `get_model_for_complexity()` method from config
- Add complexity score to plan output for user visibility
- Support upgrade workflow: `spektacular plan spec.md --detailed` forces detailed mode

## Success Metrics

### Performance Metrics
- **Cost reduction**: 70%+ reduction in token costs for simple specs
- **Speed improvement**: 50%+ faster planning for simple specs
- **Quality maintenance**: Complex spec plan quality unchanged or improved

### Accuracy Metrics
- **Classification accuracy**: 85%+ correct complexity classification vs manual review
- **User satisfaction**: <5% of users manually override complexity routing
- **Plan quality**: Fast mode plans successfully implementable 90%+ of the time

### Usage Metrics
- **Distribution**: Expect ~40% simple, ~40% medium, ~20% complex specs in real usage
- **Upgrade rate**: <15% of fast plans upgraded to detailed mode

## Non-Goals

### Out of Scope
- Dynamic model switching during agent execution (complexity determined upfront only)
- Machine learning-based complexity analysis (use rule-based approach for transparency)
- User training or onboarding for complexity system (should be transparent)
- Complexity analysis for implement command (focus on plan command only)
- Cross-spec complexity learning (each spec analyzed independently)

### Future Enhancements (Not This Version)
- Complexity analysis based on codebase size/architecture
- Team-based complexity calibration
- Historical complexity accuracy tracking and auto-tuning
- Integration with cost tracking and budgeting systems
