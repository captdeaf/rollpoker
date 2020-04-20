// Initialize sprites/etc:
var TableRenderer = {
  Templates: {
    TABLE: "#maintable",
    VIEW: "#tableview",
    HANDVIEW: "#myhandview",
  },
  Start: function() {
    $('body').html(TableRenderer.VIEW());
    $('.gamecommand').click(function(evt) {
      var me = $(this);
      var dat = me.data();
      for (var i in dat) {
        dat[i] = "" + dat[i];
      }
      console.log(me.attr('name'), dat);
      Poker.SendCommand(me.attr('name'), dat);
    });
  },
  CHIP_VALS: "",
  Update: function(data) {
    if (data.GameSettings.ChipValues != TableRenderer.CHIP_VALS) {
      TableRenderer.UpdateChips(data.GameSettings.ChipValues);
      TableRenderer.CHIP_VALS = data.GameSettings.ChipValues;
    }
    // TODO: Pick my table out from multiple tables.
    var tableData = data.Tables["table0"];
    var table = $(TableRenderer.TABLE({table:tableData, players: data.Players}));
    $('#tables').empty();
    $('#tables').append(table);
    var myp;
    _.each(data.Players, function(p) {
      if (p.PlayerId == Poker.PLAYER_ID) {
        myp = p;
      }
    });
    if (myp) {
      $('#myhand').empty();
      $('#myhand').append($(TableRenderer.HANDVIEW({player: myp})));
    }
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
      for (var i = 0; i < str.length; i += 2) {
        ret.push(str.substring(i,i+2));
      }
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
      console.log("My hand so far:" + ns);
      var m = ns.match(/m(\w+)m/);
      return m[1].match(/\w\w/g);
    }

    for (var i = 0; i < str.length; i += 2) {
      ret.push(str.substring(i,i+2));
    }
    return ret;
  },
  Setup: function() {
    for (var key in TableRenderer.Templates) {
      TableRenderer[key] = _.template($(TableRenderer.Templates[key]).html());
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
    TableRenderer.CHIP_VALUES = [];
    // Accepts "25 100 500 ..." etc.
    var colors = ["za", "zb", "zc", "zd", "ze", "zf"];
    var values = str.split(" ");
    for (var i = 0; i < values.length; i++) {
      var num = parseInt(values[i], 10);
      if (num != NaN && num > 0) {
        TableRenderer.CHIP_VALUES.push([num, colors[i]]);
      }
    }
    TableRenderer.CHIP_VALUES.sort(function(a,b) { return a[0] - b[0]; });
  },

  ChipStack: function(amt) {
    console.log("Generating chip stack for");
    console.log(amt);
    var piles = $('<div class="smallpiles">');
    for (var i = 0; amt > 0 && i < TableRenderer.CHIP_VALUES.length; i++) {
      console.log("uh");
      var pile = $('<ul class="pile">');
      var max = TableRenderer.CHIP_VALUES[i][0] * 20;
      var half = TableRenderer.CHIP_VALUES[i][0] * 10;
      var numval = amt % max;
      if (numval === 0) {
        numval = max;
      } else if (numval < half && amt > max) {
        numval += half;
      }
      amt = amt - numval;
      for (var j = 0; j < numval; j = j + TableRenderer.CHIP_VALUES[i][0]) {
        var li = $('<li>');
        li.append($('<div class="chip ' + TableRenderer.CHIP_VALUES[i][1] + '">'));
        pile.append(li);
      }
      piles.prepend(pile);
    }
    return piles.wrapAll('<div>').parent().html();
  },
};
