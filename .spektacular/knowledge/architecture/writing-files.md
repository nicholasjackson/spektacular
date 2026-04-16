# Writing Files
For reading and writing asset files such as plans, spec, or context. It can not be assumed that the agent has access to the file
system. The agent should be prompted to use the spektacular CLI to read and write these files.

When writing files in steps, the common pattern is to use the follwing template:

```
cat <<'EOF' | {{config.command}} spec goto --data '{"step":"{{next_step}}"}' --stdin spec_template
<complete filled spec here>
EOF