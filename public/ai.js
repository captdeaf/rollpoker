var PokerAI = {
  CHOICES: [
    "Call",
    "Call",
    "Call",
    "Call",
    "Call",
    "Call",
    "Call",
    "Call",
    "Call",
    "Bet",
    "Bet",
    "Bet",
    "Fold",
  ],
  Init: function() {
    Table.OnTurnStart = PokerAI.OnTurn;
    if (Poker.PLAYER.State == "TURN") {
      PokerAI.OnTurn();
    }
  },
  OnTurn: function() {
    var choice = PokerAI.CHOICES[Math.floor(Math.random() * PokerAI.CHOICES.length)];
    setTimeout(function() {
      PokerAI[choice]();
    }, Math.floor(Math.random(3000)));
  },
  Call: function() {
    if ($("input.betp").val() == "0") {
      PokerAI.SendCommand("Check", {});
    } else {
      PokerAI.SendCommand("Call", {});
    }
  },
  Bet: function() {
    PokerAI.SendCommand("Bet", {amount: "" + (parseInt($('input.betp').val()) * 2 + 50)});
  },
  Fold: function() {
    PokerAI.SendCommand("Fold", {});
  },
  SendCommand: function(cmd, args) {
    console.log("Trying:");
    console.log(cmd, args);
    Poker.SendCommand(cmd, args);
  }
};
setTimeout(PokerAI.Init, 4000);
