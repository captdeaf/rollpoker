// Initialize sprites/etc:
var TableRenderer = {
  Templates: {
    PLAYER_INFO: "#playertpl",
    TABLE: "#tabletpl",
  },
  Setup: function() {
    for (var key in this.Templates) {
      this[key] = _.template($(this.Templates[key]).html());
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

    var cards = ["A", "2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K"];
    var suits = ["S", "C", "D", "H"];
    for (var c = 0; c < cards.length; c++) {
      for (var s = 0; s < suits.length; s++) {
        makeCard(c, 0, suits[s] + cards[c]);
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

    makeChip(0, 0, "ca");
    makeChip(1, 0, "cb");
    makeChip(2, 0, "cc");
    makeChip(3, 0, "cd");
    makeChip(4, 0, "ce");
    makeChip(5, 0, "cf");
  },


  CHIP_VALUES: [],

  UpdateChips: function(str) {
    this.CHIP_VALUES = {};
    // Accepts "25 100 500 ..." etc.
    var colors = ["ca", "cb", "cc", "cd", "ce", "cf"];
    var values = str.split(" ");
    for (var i = 0; i < values.length; i++) {
      var num = parseInt(values[i], 10);
      if (num != NaN && num > 0) {
        this.CHIP_VALUES.push([num, colors[i]]);
      }
    }
    this.CHIP_VALUES.sort(function(a,b) { return a[0] - b[0]; });
  },

  UpdateSettings: function(settings) {
    this.UpdateChips(settings.ChipValues);
  },

  ChipStack: function(amt) {
    var piles = $('<div class="smallpiles">');
    for (var i = 0; amt > 0 && i < this.CHIP_VALUES.length; i++) {
      var pile = $('<ul class="pile">');
      var max = this.CHIP_VALUES[i][1] * 20;
      var half = this.CHIP_VALUES[i][1] * 10;
      var numval = amt % max;
      if (numval === 0) {
        numval = max;
      } else if (numval < half && amt > max) {
        numval += half;
      }
      amt = amt - numval;
      for (var j = 0; j < numval; j = j + this.CHIP_VALUES[i][1]) {
        var li = $('<li>');
        li.append($('<div class="chip ' + this.CHIP_VALUES[i][0] + '">'));
        pile.append(li);
      }
      piles.prepend(pile);
    }
    return piles;
  },

  UpdatePlayer: function(playerid, pinfo) {
    var piles = this.ChipStack(pinfo.chips);
    var pdiv = $(this.PLAYER_INFO(pinfo));
    var pos = $('#' + playerid);
    pos.html(pdiv);
    pos.find('.playerpiles').append(piles);
  },

  UpdatePlayerBet: function(playerid, amt) {
    var betdiv = $('#' + playerid + 'bet');
    betdiv.empty();
    var piles = chipStack(amt);
    betdiv.html(piles);
  }
};
