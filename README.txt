So with firebase.firestore, we have several documents set up:

/game/{NAME} - Private information for the server.
/public/{NAME}/table - Public information, including most of game state.
/public/{NAME}/log   - An ongoing log, both text and including event / UI
                       triggers.

/game/name Includes:
  -- Player Keys.
  -- Player Last action
  -- Table #: Deck mapping
  -- Admin password
  -- Original game settings

/public/name/table Includes:
  -- Current game state
    -- Blinds, blind schedule, and times
    -- Paused?
  -- Tables 0...:
    -- Seats 0..9:
      -- Player ID
      -- Display name
      -- Display state ("Here", "Zzzz")
      -- Current # of chips
      -- Current Rank (for busted out in Nth place)
      -- Current Hand (encrypted by player key and salt)
      -- Current bet total
      -- Current state ("waiting" "folded" etc)
    -- Pot
    -- Dealer
