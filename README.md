

- https://ollama.com/library/qwen2.5-coder:7b
- `ollama run qwen2.5-coder:7b`


go application talks to LLM
LLM talks back

-------------------

- go application talks to LLM also include the text description of what tools exists
- LLM selects tools and sends back to go application what it wants
- go application calls the tool and send back result to LLM
- LLM gets response and sends to go application
- go application relays to user

-------------------

LLM is a blackbox that takes input and gives output

self parsing tools
// self defined protocol


My Tool abstraction
// attempt to provide an abstraction


openai sdk tool calling
// one of the conventions of tool calling abstraction


mcp server
// a server that go application talks to to get the list of tools. The tools exposed by the mcp server are called mcp tools


// how does the go application select the list of tools?
// send all of them?
// match by keyword and send the "relative" ones?
// ask LLM first then filter the list?


-------------------

- go application calls mcp server and gets a list of tools (tools/list)
- go application talks to LLM also include the list of tools along with the prompt
- LLM selects tools and sends back to go application what it wants
- go application calls the tool via mcp server (tools/call) and send back result to LLM
- LLM gets response and sends to go application
- go application relays to user

-------------------

Agent Skills specification
// is a file format and directory format


there is no standard call skill protocol


has same problem as tool advertise selection, but skills are more lightweight so more could be send


-------------------

- go application loads skill handles based on Agent Skills specification
- go application sends the skills to LLM along with the message
- LLM selects skill, tells go application which one
- go application pulls out more information about the skill and send to LLM
- LLM follows skill instructions, maybe call tools and eventually sends result back to go application
- go application shows user











