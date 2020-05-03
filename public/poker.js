var CommandQueue = {
  OnTurnStart: function() {
    if (this.QueuedCommand) {
      setTimeout(function() { CommandQueue.PopCommand(); }, 1000);
    }
  },
  Queue: function(cmd, args, clearonbet) {
    this.QueuedCommand = {
      clearonbet: clearonbet,
      cmd: cmd,
      args: args,
    };
    if (Player.info.State == "TURN") {
      this.PopCommand();
    } else {
      if (args && args.amount) {
        VIEWS.Poker.Indicate("Queued: " + cmd + " " + args.amount, {canCancel: true});
      } else {
        VIEWS.Poker.Indicate("Queued: " + cmd, {canCancel: true});
      }
    }
  },
  MaybeClearQueue: function() {
    // Called on Bets. If we have a queued command that isn't Fold,
    // then cancel it.
    if (this.QueuedCommand && this.QueuedCommand.clearonbet != false) {
      this.Clear();
    }
  },
  Clear: function() {
    this.QueuedCommand = undefined;
  },
  PopCommand: function() {
    var qd = this.QueuedCommand;
    this.QueuedCommand = undefined;
    if (qd) {
      RollPoker.SendCommand(qd.cmd, qd.args);
    }
  },
};

var Render = {
  CHIP_VALUES: [
    [1, "za"],
    [25, "zb"],
    [100, "zc"],
    [500, "zd"],
    [1000, "ze"],
    [5000, "zf"],
  ],
  Card: function(cardval) {
    return '<div class="cardwrap"><div class="card ' + cardval + '"></div></div>'
  },
  Chips: function(chips, cls) {
    if (!cls) {
      cls = "betchip";
    }
    return '<div class="' + cls + '">' + chips + '</div>';
  },
  Members: function() {
    var usernames = _.map(Game.data.Members, function(name, uid) {
      if (Game.data.Hosts[uid]) {
        return name + " (" + Game.data.Hosts[uid] + ")";
      } else {
        return name;
      }
    });
    return usernames.join(", ");
  },
  ChipStack: function(amt, cls) {
    // TODO TODO: Show biggest chips possible on down to smallest.
    if (!cls) { cls = "smallpiles"; }
    var piles = $('<div class="' + cls + '">');
    for (var i = 0; amt > 0 && i < Render.CHIP_VALUES.length; i++) {
      var pile = $('<ul class="pile">');
      var max = Render.CHIP_VALUES[i][0] * 20;
      var half = Render.CHIP_VALUES[i][0] * 10;
      var numval = amt % max;
      if (numval === 0) {
        numval = max;
      } else if (numval < half && amt > max) {
        numval += half;
      }
      amt = amt - numval;
      for (var j = 0; j < numval; j = j + Render.CHIP_VALUES[i][0]) {
        var li = $('<li>');
        li.append($('<div class="chip ' + Render.CHIP_VALUES[i][1] + '">'));
        pile.append(li);
      }

      piles.prepend($('<div class="pilecontainer">').append(pile));
    }
    return piles.wrapAll('<div>').parent().html();
  },
  GetUnicodeCard: function(card) {
    var col = {
      s: "gray",
      h: "red",
      d: "red;",
      c: "gray;",
    }[card.substring(0,1)];
    var suit = {
      s: "&spades;",
      h: "&hearts;",
      d: "&diams;",
      c: "&clubs;",
    }[card.substring(0,1)];
    var cname = {
      "a": "A", "2": "2", "3": "3", "4": "4", "5": "5", "6": "6", "7": "7",
      "8": "8", "9": "9", "t": "10", "j": "J", "q": "Q", "k": "K",
    }[card.substring(1,2)];
    return '<tt style="color:' + col + '">' + cname + suit + "</tt> ";
  },
  TimeLeft: function() {
    var timeleft = Game.data.BlindTime - Math.floor((new Date()).getTime() / 1000);
    if (timeleft < 0) {
      return "Next Hand";
    }
    if (timeleft < 120) {
      return "" + timeleft + "s";
    }
    return "" + Math.round(timeleft/60) + "m";
  }
};

VIEWS.Poker = new View({
  Templates: {
    Table: "#maintable",
    View: "#tableview",
  },
  OnClick: {
    "betadd": function() {
      inp.val(parseInt(inp.val()) + Table.MINBET);
      inp.trigger("change");
    },
    "betsub": function() {
      inp.val(parseInt(inp.val()) - Table.MINBET);
      inp.trigger("change");
    },
    "cardsettings": function() {
      var menu = $("#menu");
      this.AddSubview("#menu", VIEWS.Menu);
      menu.show();
    },
    "closemenu": function() {
      this.RemoveSubview("#menu");
      $("#menu").hide();
    },
    "cancelindicate": function() {
      CommandQueue.Clear();
      this.ClearIndicator();
    },
    "cmdfold": function() {
      if (Player.info.State == "TURN") {
        CommandQueue.Queue("Fold", {}, false);
      } else {
        CommandQueue.Queue("CheckFold", {}, false);
      }
    },
    "cmdcall": function() {
      CommandQueue.Queue("Call", {}, true);
    },
    "cmdbet": function() {
      CommandQueue.Queue("Bet", {amount: "" + (this.Bet + (Player.Table.CurBet - Player.info.Bet))}, true);
    },
    "addbb": function() {
      this.WannaBet(this.Bet + Player.Table.Blinds[Player.Table.Blinds.length - 1]);
    },
    "subbb": function() {
      this.WannaBet(this.Bet - Player.Table.Blinds[Player.Table.Blinds.length - 1]);
    },
    "subpot": function() {
      this.WannaBet(this.Bet - Player.Table.Pot);
    },
    "addpot": function() {
      this.WannaBet(this.Bet + Player.Table.Pot);
    },
  },
  OnChange: {
    "betval": function(evt) {
      this.WannaBet(parseInt(evt.target.value, 10));
    }
  },
  Start: function() {
    $('body').html(this.T.View());
  },
  OnTurnStart: function() {
    if (window.Notification && window.Notification.permission == "granted") {
      var notif = new window.Notification("Your turn");
      notif.onclick = function() {
        window.focus();
        notif.close();
      }
      setTimeout(function() {notif.close();}, 4000);
    }
    CommandQueue.OnTurnStart();
  },
  Update: function(data) {
    // TODO: Pick my table out from multiple tables.
    Player.Table = data.Tables["table0"];
    var table = $(this.T.Table({game: data, table:Player.Table, players: data.Players}));
    $('#mytable').empty();
    $('#mytable').append(table);
    if (_.any(Player.info)) {
      this.UpdateBetButtons();
      if (_.any(Player.pdata.Cards)) {
        if (this.ValueDiffers("hand", Player.pdata.Cards)) {
          this.ClearIndicator();
        }
        if (this.ValueDiffers("State", Player.info.State)) {
          if (Player.info.State == "TURN") {
            this.OnTurnStart();
            this.ClearIndicator();
          }
        }
        if (this.ValueDiffers("minbet", Player.Table.MinBet)) {
          this.WannaBet(Player.Table.MinBet);
        }
        $('#myhand').show();
        $('#betplaque').show();
      } else {
        // I folded or busted out.
        $('#myhand').hide();
        $('#betplaque').hide();
      }
    }
  },
  LASTHAND: [],
  Bet: 0,
  GetBetAmount: function(amt, maxbet) {
    if (maxbet <= amt) {
      // Can't bet higher than the chips you have
      return maxbet;
    }
    if (amt > maxbet) {
      // Can't bet more than you have
      return maxbet;
    }
    if (maxbet < Player.Table.MinBet) {
      // Can't bet lower than MinBet unless you're all-in.
      return maxbet;
    }
    if (amt < Player.Table.MinBet) {
      return Player.Table.MinBet;
    }
    return amt;
  },
  UpdateBetButtons: function() {
    var curbet = Player.Table.CurBet;
    var minbet = Player.Table.MinBet;
    var playerbet = Player.info.Bet;
    var diffbet = curbet - playerbet;
    var state = Player.info.State;
    var foldb = $("#cmdfold");
    var callb = $("#cmdcall");
    var betb = $("#cmdbet");
    $("#addbb").text("+ " + minbet);
    $("#subbb").text("- " + minbet);
    console.log(minbet, curbet, playerbet, diffbet);
    if (state == "FOLDED") {
      $(".gameinput").prop("disabled", true);
      // Player's out.
      foldb.text("Folded");
      callb.text("Check");
      betb.text("Bet");
      $("#betplaque").find("button").prop('disabled', true);
      return;
    } else if (Player.info.Chips < 1) {
      $(".gameinput").prop("disabled", true);
      foldb.text("All-In");
      callb.text("All-In");
      betb.text("All-In");
    } else {
      $(".gameinput").prop("disabled", false);
      foldb.prop('disabled', false);
      callb.prop('disabled', false);
      betb.prop('disabled', false);
      foldb.text("Fold");
      if (diffbet > 0) {
        callb.text("Call");
      } else {
        callb.text("Check");
      }
      if (curbet > 0) {
        betb.text("Raise");
      } else {
        betb.text("Bet");
      }
      $("#betplaque").find("button").prop('disabled', false);
    }
    // If Call will put a player all-in, disable bet.
    if (diffbet >= Player.info.Chips) {
      betb.prop('disabled', true);
      $("#betplaque").find("button").prop('disabled', true);
      callb.text("All-In Call");
    }
    if (state != "TURN") {
      // It's not player's turn: These buttons are for queueing commands.
      if (diffbet == 0) {
        foldb.text("Check/Fold");
      }
    }
  },
  WannaBet: function(amt) {
    var maxbet = Player.info.Chips - (Player.Table.CurBet - Player.info.Bet);
    this.Bet = this.GetBetAmount(amt, maxbet);
    $("#betval").val(this.Bet)
    if (this.Bet == maxbet) {
      console.log("All In");
      // ALL IN
    }
  },
  OnLog: function(message) {
    var upd = message.replace(/<<(\w+)>>/g, function(a, b) {
      return Render.GetUnicodeCard(b);
    });
    var ls = $('#logscreen');
    if (upd == "New Hand") {
      upd = "<hr>";
    }
    ls.append($("<p>").html(upd));
    ls[0].scrollTop = ls[0].scrollHeight;
  },
  GetPlayerLocation: function(playerid) {
    return $("#" + playerid).offset();
  },
  INDICATING: undefined,
  Indicate: function(message, opts) {
    if (opts && opts.canCancel) {
      $("#cancelindicate").show();
    } else {
      $("#cancelindicate").hide();
    }
    $("#indication").text(message);
    $("#indicatorbar").show();
    $(".gameinput").prop("disabled", true);
    this.INDICATING = message;
  },
  ClearIndicator: function() {
    this.INDICATING = undefined;
    $("#indicatorbar").hide();
    this.UpdateBetButtons();
  },
  OnSecond: function() {
    $("#timeleft").text(Render.TimeLeft());
  },
  OnEvent: {
    StartBets: function() {
      CommandQueue.Clear();
      this.ClearIndicator();
    },
    Call: function(playerid, amt, opt) {
      var off = this.GetPlayerLocation(playerid);
      if (off) {
        if (opt == "ALL-IN") {
          Helpers.RaiseImage(off, "Allin");
        } else if (amt == 0) {
          Helpers.RaiseImage(off, "Check");
        } else {
          Helpers.RaiseImage(off, "Call");
        }
      }
    },
    Fold: function(playerid) {
      var off = this.GetPlayerLocation(playerid);
      if (off) {
        Helpers.RaiseImage(off, "Folded");
      }
    },
    Bet: function(playerid, amt, opt) {
      var off = this.GetPlayerLocation(playerid);
      CommandQueue.MaybeClearQueue();
      if (off) {
        if (opt == "ALL-IN") {
          Helpers.RaiseImage(off, "Allin");
        } else if (amt != Player.Table.CurBet) {
          Helpers.RaiseImage(off, "Raise");
        } else {
          Helpers.RaiseImage(off, "Bet");
        }
      }
    },
    Win: function(playerid, amt, handname, handcards) {
      var off = this.GetPlayerLocation(playerid);
      if (off) {
        Helpers.RaiseText(off, "Won " + amt);
      }
    }
  },
  RaiseTextFor: function(playerid, text, cls) {
    var off = this.GetPlayerLocation(playerid);
    Helpers.RaiseText(off, text, cls);
  }
});
