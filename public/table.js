// Initialize sprites/etc:
var Table = {
  Templates: {
    TABLE: "#maintable",
    VIEW: "#tableview",
    HANDVIEW: "#myhandview",
    BETVIEW: "#betplaqueview",
    MENU: "#menucontents",
  },
  QueuedCommand: undefined,
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
  Start: function(doc) {
    $('body').html(Table.VIEW());
    if (Poker.PLAYER) {
      function canStart() {
        if (Table.QueuedCommand) return false;
        return Poker.PLAYER.State == "TURN" || Poker.PLAYER.State == "WAITING"
      }
      var lastDrop = 0;
      Table.SetDraggable($('#myhand'), {vert: true, canStart: canStart}, function(el, data) {
        if (data.yd < -150) {
          // Swiped the cards up. Fold.
          Table.QueueCommand("Fold", {}, false);
        } else if (data.yd < 4 && data.yd > -4 && data.xd > -4 && data.xd < 4) {
          if ((Date.now() - lastDrop) < 700) {
            Table.QueueCommand("Check", {}, true);
          } else {
            lastDrop = Date.now();
          }
        }
      });
      var bp = $('#betplaque');
      bp.empty();
      bp.append($(Table.BETVIEW({player: Poker.PLAYER})));
      Table.SetDraggable($('#betplaque'), {vert: true, canStart: canStart}, function(el, data) {
        if (data.yd < -150) {
          console.log("Trying to bet");
          // Swiped the bet up to table area. Call, Bet or Raise.
          // Force string of int.
          var bet = "" + parseInt(inp.val());
          Table.QueueCommand("Bet", {amount: bet}, true);
        }
      });
      var inp = bp.find('input.betp');
      var chipv = bp.find('.betchips');
      var tt = bp.find('.betdesc');
      inp.change(function(val) {
        var min = Table.CURBET - Poker.PLAYER.Bet;
        inp.val(parseInt(inp.val()));
        if (inp.val() < min) {
          inp.val(min);
        }
        if (inp.val() > Poker.PLAYER.Chips) {
          inp.val(Poker.PLAYER.Chips);
        }
        if (inp.val() == 0) {
            tt.text("Check");
        }
        if (inp.val() == min && Table.CURBET != 0) {
            tt.text("Call");
        }
        chipv.html(Table.ChipStack(inp.val(), "bigpiles"));
        if (inp.val() > (Table.CURBET - Poker.PLAYER.Bet)) {
          if (Table.CURBET == 0) {
            tt.text("Bet " + inp.val());
          } else {
            tt.text("Raise " + (inp.val() - Table.CURBET));
          }
        }
        if (Poker.PLAYER.Chips == inp.val()) {
          tt.text("All-In");
        }
      });
      bp.find('button[name="betcall"]').on("click touchstart", function() {
        inp.val(Table.CURBET - Poker.PLAYER.Bet);
        inp.trigger("change");
      });
      bp.find('button[name="betadd"]').on("click touchstart", function() {
        inp.val(parseInt(inp.val()) + Table.MINBET);
        inp.trigger("change");
      });
      bp.find('button[name="betsub"]').on("click touchstart", function() {
        inp.val(parseInt(inp.val()) - Table.MINBET);
        inp.trigger("change");
      });
      inp.on("click touchstart", function() {
        inp.focus();
      });
      $("#cancelindicate").on("click touchstart", function() {
        Table.QueueCancel();
        Table.ClearIndicator();
      });
      $("#cardsettings").on("click touchstart", Table.ShowMenu);
    }
    Poker.LogCallback = Table.LogUpdate;
  },
  ShowMenu: function() {
    var menu = $("#menu");
    menu.html(Table.MENU({data: Poker.DATA, player: Poker.PLAYER}));
    menu.find("#notifications").on("click touchstart", function() {
      Notification.requestPermission();
    });
    menu.find("#closemenu").on("click touchstart", function() {
      menu.hide();
    });
    menu.show();
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
    if (Table.QueuedCommand) {
      setTimeout(function() { Table.PopCommand(); }, 1000);
    }
  },
  CHIP_VALS: "",
  CUR_BET: -1,
  Update: function(data) {
    if (data.GameSettings.ChipValues != Table.CHIP_VALS) {
      Table.UpdateChips(data.GameSettings.ChipValues);
      Table.CHIP_VALS = data.GameSettings.ChipValues;
    }
    // TODO: Pick my table out from multiple tables.
    var tableData = data.Tables["table0"];
    var table = $(Table.TABLE({game: data, table:tableData, players: data.Players}));
    $('#mytable').empty();
    $('#mytable').append(table);
    Table.ShowPlayerState(tableData,Poker.PLAYER);
    if (Poker.PLAYER) {
      if (Poker.PLAYER.Hand && Poker.PLAYER.Hand.length > 0) {
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
      $('button[name="betcall"]').text("= " + (Table.CURBET - player.Bet) + " (Call)");
      $('button[name="betadd"]').text("+ " + Table.MINBET);
      $('button[name="betsub"]').text("- " + Table.MINBET);
      var inp = $('input.betp');
      inp.val(tableData.CurBet - Poker.PLAYER.Bet);
      var callbutt = $('button[name="betcall"]');
      callbutt.text((tableData.CurBet - Poker.PLAYER.Bet) + " (call)");
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
  GetHand: function(str) {
    var ret = [];
    if (str.startsWith("!")) {
      str = atob(str.substring(1,str.length))
      // Figure out how many cards. There's 38 chars in encryption
      // !, a-z0-9, an extra m, then the cards (2 chars each)
      for (var i = 38; i < str.length; i += 2) {
        ret.push("bg");
      }
    } else {
      return str.match(/(\w\w)/g);
    }
    return ret
  },
  DecryptMyHand: function(str) {
    var ret = [];
    if (str.startsWith("!")) {
      var ns = "";
      var bstr = atob(str.substring(1,str.length));
      // Decrypt: De-XOR it, then pluck the string out from m.*m
      for (var i = 0; i < bstr.length; i++) {
        ns = ns + String.fromCharCode(bstr.charCodeAt(i) ^ Poker.PLAYER_KEY.charCodeAt(i % Poker.PLAYER_KEY.length));
      }
      var m = ns.match(/m(\w+)m/);
      return m[1].match(/\w\w/g);
    }

    for (var i = 0; i < str.length; i += 2) {
      ret.push(str.substring(i,i+2));
    }
    return ret;
  },
  Setup: function() {
    for (var key in Table.Templates) {
      Table[key] = _.template($(Table.Templates[key]).html());
    }

    function makeCard(x,y,cardname) {
      var style = document.createElement('style');
      style.type = 'text/css';
      var cardx = x * 61.5;
      var cardy = y * 81;
      style.innerHTML = "." + cardname +
        "{ background-position: -" + cardx + "px -" + cardy + "px; }";
      document.getElementsByTagName('head')[0].appendChild(style);
    }

    var cards = ["a", "2", "3", "4", "5", "6", "7", "8", "9", "t", "j", "q", "k"];
    var suits = ["s", "c", "d", "h"];
    for (var c = 0; c < cards.length; c++) {
      for (var s = 0; s < suits.length; s++) {
        makeCard(c, s, suits[s] + cards[c]);
      }
    }

    function makeChip(x,y,chipname) {
      var style = document.createElement('style');
      style.type = 'text/css';
      var chipx = x * 129;
      var chipy = y * 59;
      style.innerHTML = "." + chipname +
        "{ background-position: -" + chipx + "px -" + chipy + "px; }";
      document.getElementsByTagName('head')[0].appendChild(style);
    }

    makeChip(0, 0, "za");
    makeChip(1, 0, "zb");
    makeChip(2, 0, "zc");
    makeChip(3, 0, "zd");
    makeChip(4, 0, "ze");
    makeChip(5, 0, "zf");
  },


  CHIP_VALUES: [],

  UpdateChips: function(str) {
    Table.CHIP_VALUES = [];
    // Accepts "25 100 500 ..." etc.
    var colors = ["za", "zb", "zc", "zd", "ze", "zf"];
    var values = str.split(" ");
    for (var i = 0; i < values.length; i++) {
      var num = parseInt(values[i], 10);
      if (num != NaN && num > 0) {
        Table.CHIP_VALUES.push([num, colors[i]]);
      }
    }
    Table.CHIP_VALUES.sort(function(a,b) { return a[0] - b[0]; });
  },

  ChipStack: function(amt, cls) {
    if (!cls) { cls = "smallpiles"; }
    var piles = $('<div class="' + cls + '">');
    for (var i = 0; amt > 0 && i < Table.CHIP_VALUES.length; i++) {
      var pile = $('<ul class="pile">');
      var max = Table.CHIP_VALUES[i][0] * 20;
      var half = Table.CHIP_VALUES[i][0] * 10;
      var numval = amt % max;
      if (numval === 0) {
        numval = max;
      } else if (numval < half && amt > max) {
        numval += half;
      }
      amt = amt - numval;
      for (var j = 0; j < numval; j = j + Table.CHIP_VALUES[i][0]) {
        var li = $('<li>');
        li.append($('<div class="chip ' + Table.CHIP_VALUES[i][1] + '">'));
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
      "a": "1",
      "2": "2",
      "3": "3",
      "4": "4",
      "5": "5",
      "6": "6",
      "7": "7",
      "8": "8",
      "9": "9",
      "t": "a",
      "j": "b",
      "q": "d",
      "k": "e",
    }[card.substring(1,2)];
    return pref + suff;
  },
  LogUpdate: function(ts, message) {
    var upd = message.replace(/<<(\w+)>>/g, function(a, b) {
      return Table.GetUnicodeCard(b);
    });
    var ls = $('#logscreen');
    ls.append($("<p>").html(upd));
    ls[0].scrollTop = ls[0].scrollHeight;
  },
};
