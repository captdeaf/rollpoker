// Initialize sprites/etc:
function onBodyLoad() {
  function makeCard(x,y,cardname) {
    var style = document.createElement('style');
    style.type = 'text/css';
    var cardx = x * 61.5;
    var cardy = y * 81;
    style.innerHTML = "." + cardname +
      "{ background-position: -" + cardx + "px -" + cardy + "px; }";
    document.getElementsByTagName('head')[0].appendChild(style);
  }

  for (var i = 0 ; i < 13; i++) {
    makeCard(i, 0, "s" + (i+1));
    makeCard(i, 1, "c" + (i+1));
    makeCard(i, 2, "d" + (i+1));
    makeCard(i, 3, "h" + (i+1));
  }

  var chipValues = [];

  function makeChip(x,y,chipname,value) {
    var style = document.createElement('style');
    style.type = 'text/css';
    var chipx = x * 129;
    var chipy = y * 59;
    chipValues.push([chipname,value]);
    style.innerHTML = "." + chipname +
      "{ background-position: -" + chipx + "px -" + chipy + "px; }";
    document.getElementsByTagName('head')[0].appendChild(style);
  }

  makeChip(0, 0, "chipwhite", 1);
  makeChip(1, 0, "chipred", 5);
  makeChip(2, 0, "chipblue", 25);
  makeChip(3, 0, "chipgreen", 100);
  makeChip(4, 0, "chipblack", 500);
  makeChip(5, 0, "chipyellow", 1000);

  function chipStack(amt) {
    var piles = $('<div class="smallpiles">');
    for (var i = 0; amt > 0 && i < chipValues.length; i++) {
      var pile = $('<ul class="pile">');
      var max = chipValues[i][1] * 20;
      var half = chipValues[i][1] * 10;
      var numval = amt % max;
      if (numval === 0) {
        numval = max;
      } else if (numval < half && amt > max) {
        numval += half;
      }
      amt = amt - numval;
      for (var j = 0; j < numval; j = j + chipValues[i][1]) {
        var li = $('<li>');
        li.append($('<div class="chip ' + chipValues[i][0] + '">'));
        pile.append(li);
      }
      piles.prepend(pile);
    }
    return piles;
  }

  var playerInfoTemplate = _.template($('#playertpl').html());
  function populatePlayerInfo(playerid, pinfo) {
    var piles = chipStack(pinfo.chips);
    var pdiv = $(playerInfoTemplate(pinfo));
    var pos = $('#' + playerid);
    pos.html(pdiv);
    pos.find('.playerpiles').append(piles);
  }

  function populateBet(playerid, amt) {
    var betdiv = $('#' + playerid + 'bet');
    betdiv.empty();
    var piles = chipStack(amt);
    betdiv.html(piles);
  }

  populatePlayerInfo("player1", {
    name: "Prawn",
    chips: 1500,
    state: "Folded",
    cards: ["h13","h12"],
  });

  populatePlayerInfo("player2", {
    name: "Rob",
    chips: 1500,
    state: "All In",
    cards: ["h2","s7"],
  });

  populatePlayerInfo("player4", {
    name: "Greg",
    chips: 1500,
    state: "Folded",
    cards: ["cardbg","cardbg"],
  });

  populatePlayerInfo("player6", {
    name: "Nick",
    chips: 1500,
    state: "Lunatic",
    cards: ["cardbg","cardbg"],
  });

  populatePlayerInfo("player8", {
    name: "Sank",
    chips: 1500,
    state: "Folded",
    cards: ["cardbg","cardbg"],
  });

  populatePlayerInfo("player10", {
    name: "Nesmith",
    chips: 1500,
    state: "Folded",
    cards: ["cardbg","cardbg"],
  });

  for (var i = 1; i <= 10; i++) {
    populateBet("player" + i, 1834);
  }
}
