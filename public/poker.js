var CommandQueue = {
  OnTurnStart: function() {
    if (window.Notification && window.Notification.permission == "granted") {
      var notif = new window.Notification("Your turn");
      notif.onclick = function() {
        window.focus();
        notif.close();
      }
      setTimeout(function() {notif.close();}, 4000);
    }
    if (Table.QueuedCommand) {
      setTimeout(function() { Table.PopCommand(); }, 1000);
    }
  },
  QueueCommand: function(cmd, args, clearonbet) {
    Table.QueuedCommand = {
      clearonbet: clearonbet,
      cmd: cmd,
      args: args,
    };
    if (Poker.PLAYER.State == "TURN") {
      Table.PopCommand();
    } else {
      Table.Indicate("Queued: " + cmd, {canCancel: true});
    }
  },
  MaybeClearQueue: function() {
    // Called on Bets. If we have a queued command that isn't Fold,
    // then cancel it.
    if (Table.QueuedCommand && Table.QueuedCommand.cmd != "Fold") {
      Table.QueueCancel();
    }
  },
  QueueCancel: function() {
    Table.QueuedCommand = undefined;
    Table.ClearIndicator();
  },
};

var Indicator = {
  INDICATING: undefined,
  Indicate: function(message, opts) {
    if (opts && opts.canCancel) {
      $("#cancelindicate").show();
    } else {
      $("#cancelindicate").hide();
    }
    $("#indication").text(message);
    $("#indicatorbar").show();
    Table.INDICATING = message;
  },
  ClearIndicator: function() {
    Table.INDICATING = undefined;
    $("#indicatorbar").hide();
  },
  PopCommand: function() {
    if (Table.INDICATING) {
      Table.ClearIndicator();
    }
    var qd = Table.QueuedCommand;
    Table.QueuedCommand = undefined;
    Poker.SendCommand(qd.cmd, qd.args);
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
    var pref = {
      s: "&#x1f0a",
      h: "&#x1f0b",
      d: "&#x1f0c",
      c: "&#x1f0d",
    }[card.substring(0,1)];
    var suff = {
      "a": "1", "2": "2", "3": "3", "4": "4", "5": "5", "6": "6", "7": "7",
      "8": "8", "9": "9", "t": "a", "j": "b", "q": "d", "k": "e",
    }[card.substring(1,2)];
    return pref + suff;
  },
};

VIEWS.Poker = new View({
  Templates: {
    Table: "#maintable",
    View: "#tableview",
    Hand: "#myhandview",
    Bet: "#betplaqueview",
    Menu: "#menucontents",
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
      menu.html(Table.MENU({data: Game.data, player: Player.info}));
      menu.show();
    },
    "closemenu": function() {
      $("#menu").hide();
    },
    "asknotifications": function() {
      Notification.requestPermission();
    },
    "cancelindicate": function() {
      CommandQueue.QueueCancel();
      this.ClearIndicator();
    },
  },
  Start: function() {
    $('body').html(this.T.View());
  },
  CHIP_VALS: "",
  CUR_BET: -1,
  Update: function(data) {
    // TODO: Pick my table out from multiple tables.
    var tableData = data.Tables["table0"];
    var table = $(this.T.Table({game: data, table:tableData, players: data.Players}));
    $('#mytable').empty();
    $('#mytable').append(table);
    this.ShowPlayerState(tableData,Player.info);
    if (_.any(Player.info)) {
      if (_.any(Player.info.Hand)) {
        console.log("State: " + Poker.PLAYER.State);
        if (!Table.IsSameHand(Poker.PLAYER.Hand, Table.LASTHAND)) {
          Table.CUR_BET = -1;
          Table.LASTHAND = Poker.PLAYER.Hand;
          Table.UpdateHand(Poker.PLAYER);
          Table.ClearIndicator();
        }
        if (Poker.PLAYER.State != Table.LAST_PLAYER_STATE) {
          if (Poker.PLAYER.State == "TURN") {
            console.log("Trying OnTurnStart");
            Table.OnTurnStart();
          } else {
            Table.ClearIndicator();
          }
        }
        if (tableData.CurBet != Table.CUR_BET) {
          Table.CUR_BET = tableData.CurBet;
          Table.UpdateBetPlaque(tableData,Poker.PLAYER);
        }
        Table.ShowPlayerState(tableData,Poker.PLAYER);
        $('#myhand').show();
        $('#betplaque').show();
      } else {
        $('#myhand').hide();
        $('#betplaque').hide();
        // Table.Indicate("You Folded");
      }
      Table.LAST_PLAYER_STATE = Poker.PLAYER.State
    }
  },
  ShowPlayerState: function(table, player) {
    $('#indicator').text(Table.GetIndicatorText(table, player));
  },
  GetIndicatorText: function(table, player) {
    if (!player) return "Spectating";
    if (player.State == "TURN") {
      return "Your Turn";
    } else {
      return player.State;
    }
  },
  LASTHAND: [],
  IsSameHand: function(ary, ary2) {
    if (ary.length != ary2) return false;
    for (var i = 0; i < ary.length; i++) {
      if (ary[i] != ary2) return false;
    }
    return true
  },
  MINBET: -1,
  CURBET: -1,
  PLYBET: -1,
  UpdateBetPlaque: function(tableData, player, val) {
    if (Table.MINBET != tableData.MinBet || Table.CURBET != tableData.CurBet ||
        player.Bet != Table.PLYBET) {
      Table.MINBET = tableData.MinBet;
      Table.CURBET = tableData.CurBet;
      Table.PLYBET = player.Bet;
      $('button[name="betadd"]').text("+ " + Table.MINBET);
      $('button[name="betsub"]').text("- " + Table.MINBET);
      var inp = $('input.betp');
      inp.val(tableData.CurBet - Poker.PLAYER.Bet);
      var callbutt = $('button[name="betcall"]');
      if (tableData.CurBet - Poker.PLAYER.Bet == 0) {
        callbutt.text("0 (Check)");
      } else {
        callbutt.text((tableData.CurBet - Poker.PLAYER.Bet) + " (Call)");
      }
      inp.trigger("change");
    }
  },
  UpdateHand: function(player) {
    $('#myhand').empty();
    if (player.Hand.length > 0) {
      $('#myhand').append($(Table.HANDVIEW({player: player})));
    }
  },
  SetDraggable: function(jqe, opts, cb, cbpress) {
    function handle_mousedown(e){
      if (window.dragging != undefined) return;
      if (opts.canStart) {
        if (!opts.canStart()) return;
      }
      window.dragging = true;
      var origX = e.pageX;
      var origY = e.pageY;
      var off = jqe.offset();
      function handle_dragging(evt){
        var newoff = {
          left: off.left,
          top: off.top,
        }
        if (opts.horiz) newoff.left += (evt.pageX - origX);
        if (opts.vert) newoff.top += (evt.pageY - origY);
        jqe.offset(newoff);
        evt.preventDefault();
      }
      function handle_mouseup(evt){
        if (cbpress) {
          cbpress("end");
        }
        $('body')
        .off('touchend', handle_mouseup)
        .off('touchmove', handle_dragging)
        .off('touchcancel', handle_mouseup)
        .off('mousemove', handle_dragging)
        .off('mouseup', handle_mouseup)
        .off('mouseleave', handle_mouseup);
        jqe.offset(off);
        cb(jqe, {x: evt.pageX, y: evt.pageY, xd: evt.pageX - origX, yd: evt.pageY - origY});
        evt.preventDefault();
        window.dragging = undefined;
      }
      $('body')
      .on('mouseleave', handle_mouseup)
      .on('mouseup', handle_mouseup)
      .on('mousemove', handle_dragging)
      .on('touchend', handle_mouseup)
      .on('touchmove', handle_dragging)
      .on('touchcancel', handle_mouseup);
      if (cbpress) {
        cbpress("start");
      }
      e.preventDefault();
    }
    jqe.mousedown(handle_mousedown);
    jqe.on("touchstart", handle_mousedown);
  },
  Setup: function() {
  },

  OnLog: function(message) {
    var upd = message.replace(/<<(\w+)>>/g, function(a, b) {
      return Table.GetUnicodeCard(b);
    });
    var ls = $('#logscreen');
    ls.append($("<p>").html(upd));
    ls[0].scrollTop = ls[0].scrollHeight;
  },
});
