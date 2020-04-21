// Initialize sprites/etc:
var Table = {
  Templates: {
    TABLE: "#maintable",
    VIEW: "#tableview",
    HANDVIEW: "#myhandview",
    BETVIEW: "#betplaqueview",
  },
  Start: function(doc) {
    $('body').html(Table.VIEW());
    $('.gamecommand').click(function(evt) {
      var me = $(this);
      var dat = me.data();
      for (var i in dat) {
        dat[i] = "" + dat[i];
      }
      console.log(me.attr('name'), dat);
      Poker.SendCommand(me.attr('name'), dat);
    });
    if (Poker.PLAYER) {
      var lastDrop = 0;
      Table.SetDraggable($('#myhand'), {vert: true}, function(el, data) {
        console.log("Dropped:");
        console.log(data);
        if (data.yd < -150) {
          // Swiped the cards up. Fold.
          Poker.SendCommand("Fold", {});
        } else if ((Date.now() - lastDrop) < 1500) {
          Poker.SendCommand("Check", {});
        } else {
          lastDrop = Date.now();
        }
      });
      var bp = $('#betplaque');
      bp.empty();
      bp.append($(Table.BETVIEW({player: Poker.PLAYER})));
      Table.SetDraggable($('#betplaque'), {vert: true}, function(el, data) {
        if (data.yd < -150) {
          console.log("Trying to bet");
          // Swiped the bet up to table area. Call, Bet or Raise.
          // Force string of int.
          var bet = "" + parseInt(inp.val());
          Poker.SendCommand("Bet", {amount: bet});
        }
      });
      var inp = bp.find('input.betp');
      var chipv = bp.find('.betchips');
      var tt = bp.find('.betdesc');
      inp.change(function(val) {
        console.log("Huh?");
        
        var min = Table.CURBET - Poker.PLAYER.Bet;
        inp.val(parseInt(inp.val()));
        if (inp.val() < min) {
          inp.val(min);
        }
        if (inp.val() > Poker.PLAYER.Chips) {
          inp.val(Poker.PLAYER.Chips);
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
      bp.find('button[name="betcall"]').click(function() {
        inp.val(Table.CURBET - Poker.PLAYER.Bet);
        inp.trigger("change");
      });
      bp.find('button[name="betadd"]').click(function() {
        inp.val(parseInt(inp.val()) + Table.MINBET);
        inp.trigger("change");
      });
      bp.find('button[name="betsub"]').click(function() {
        inp.val(parseInt(inp.val()) - Table.MINBET);
        inp.trigger("change");
      });
    }
  },
  CHIP_VALS: "",
  Update: function(data) {
    if (data.GameSettings.ChipValues != Table.CHIP_VALS) {
      Table.UpdateChips(data.GameSettings.ChipValues);
      Table.CHIP_VALS = data.GameSettings.ChipValues;
    }
    // TODO: Pick my table out from multiple tables.
    var tableData = data.Tables["table0"];
    var table = $(Table.TABLE({table:tableData, players: data.Players}));
    $('#tables').empty();
    $('#tables').append(table);
    if (Poker.PLAYER) {
      if (!Table.IsSameHand(Poker.PLAYER.Hand, Table.LASTHAND)) {
        Table.LASTHAND = Poker.PLAYER.Hand;
        Table.UpdateHand(Poker.PLAYER);
      }
      Table.UpdateBetPlaque(tableData,Poker.PLAYER);
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
        Player.Bet != Table.PLYBET) {
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
  SetDraggable: function(jqe, opts, cb) {
    function handle_mousedown(e){
      if (window.dragging != undefined) return;
      window.dragging = true;
      var origX = e.pageX;
      var origY = e.pageY;
      var off = jqe.offset();
      function handle_dragging(e){
        var newoff = {
          left: off.left,
          top: off.top,
        }
        if (opts.horiz) newoff.left += (e.pageX - origX);
        if (opts.vert) newoff.top += (e.pageY - origY);
        jqe.offset(newoff);
      }
      function handle_mouseup(e){
        $('body')
        .off('mousemove', handle_dragging)
        .off('mouseup', handle_mouseup)
        .off('mouseleave', handle_mouseup);
        jqe.offset(off);
        cb(jqe, {x: e.pageX, y: e.pageY, xd: e.pageX - origX, yd: e.pageY - origY});
        window.dragging = undefined;
      }
      $('body')
      .on('mouseleave', handle_mouseup)
      .on('mouseup', handle_mouseup)
      .on('mousemove', handle_dragging);
    }
    jqe.mousedown(handle_mousedown);
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
};
