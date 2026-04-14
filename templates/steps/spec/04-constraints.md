## Step {{step}}: {{title}}

Ask the user: Are there any hard constraints or boundaries the solution must operate within?

Examples:
• Must integrate with the existing authentication system
• Cannot introduce breaking changes to the public API
• Must support the current minimum supported runtime versions

**Constraints are boundaries, not design decisions.** Apply this test to every constraint before you write it down:

> *If this constraint were removed, would the feature become impossible, or just implemented differently?*

Only the first kind belongs here. The second kind is a design decision and belongs in Technical Approach.

• Real constraint: *"must not break the shape of the public JSON response"* — removing it lets the feature break downstream consumers.
• Real constraint: *"must not require new config keys"* — removing it changes the deployment contract.
• Not a constraint: *"must use the existing FSM engine"* — removing it just lets you pick a different mechanism. That's a design choice; move it to Technical Approach.
• Not a constraint: *"must use markdown templates with mustache rendering"* — same reason.

Constraints are usually phrased as "must not break X" or "must stay compatible with Y", not "must use Z". If the user gives you a "must use" item, ask whether removing it would make the feature impossible — if not, park it for Technical Approach.

Capture their response. If blank, note that there are no constraints.

Once you are satisfied, move to the next step by running the command:

{{config.command}} spec goto --data '{"step":"{{next_step}}"}'

