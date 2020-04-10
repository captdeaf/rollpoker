Describes the JSON Object encapsulating an entire game.

Document Index: // The game name, auto-generated, with settings.
{
  Settings: {}, // All settings. For, e.g: Game type, chip types, starting
                // counts, pot limit, no limit, etc.
  Players: {
    s1: { // Seat #, 0-9
      email: "address@foo.com", // For sending reconnect links and logs.
      display name: "", // "Greg", "Sank", "Prawn", etc.
      chips: 1234, // Current chip count
      hand: ["spade3", "heart9"], // Dealt cards. Filtered by player, replaced with "cardbg"s
      state: 'name', // "folded", "bet", "allin", "sitting out", "needscall", "called", etc.
    },
    3: {...},
    ...
  },
  gamestate: { // Contains current state info
    blinds: [],
    blindstate: "paused", "running", etc.
    blindtime: <timestamp in seconds of last blind event (raise, pause, etc)>,
    pot: 123,
  },
  deck: [], // Current state of deck, post-shuffling/etc.
  events: [ // A running log of past 20 "events". Anybody missing more than 20 gets full state.
    {id: 123, name:"burn"},
    {id: 124, name:"flop"},
    {id: 125, name:"waitbet",player:"1",min:0,bb:20,max:300},
    {id: 126, name:"bet",player:"1",amount:50},
    {id: 127, name:"waitbet",player:"3"},
    ...
  ],
  log: [ // Running log of text descriptions.
    "Chris folded",
    "Greg raised to C500",
  ],
}

