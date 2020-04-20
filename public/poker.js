
var Poker = {
  SetupFirebase: function() {
    // Your web app's Firebase configuration
    var firebaseConfig = {
      apiKey: "AIzaSyAsTsJ7UjBQu8CMADJP-JFysn6ON8Hm77M",
      authDomain: "rollpoker.firebaseapp.com",
      databaseURL: "https://rollpoker.firebaseio.com",
      projectId: "rollpoker",
      storageBucket: "rollpoker.appspot.com",
      messagingSenderId: "413322307823",
      appId: "1:413322307823:web:2d12f3485f45d55b12d31a"
    };
    // Initialize Firebase
    firebase.initializeApp(firebaseConfig);
  },
  Setup: function() {
    // First make sure we have our game name.
    var m = document.location.pathname.match(/table\/(\w+)$/);
    if (m) {
      Poker.NAME = m[1];
    } else {
      return;
    }
    var m = document.location.search.match(/\?id=(\w+)\&key=(\w+)$/);
    if (m) {
      Poker.PLAYER_ID = m[1];
      Poker.PLAYER_KEY = m[2];
      Poker.SetPlayerCookie("playerid", Poker.PLAYER_ID);
      Poker.SetPlayerCookie("playerkey", Poker.PLAYER_KEY);
      document.location.search = "";
    }
    if (!Poker.PLAYER_ID) {
      Poker.PLAYER_ID = Poker.GetPlayerId();
      Poker.PLAYER_KEY = Poker.GetPlayerKey();
    }

    // Initialize all the renderers
    Signup.Setup();
    TableRenderer.Setup();
  },
  SetPlayerCookie: function(name, val) {
    // We set this in a cookie.
    var d = new Date();
    d.setTime(d.getTime() + (365*24*60*60*1000));
    var expires = "expires="+ d.toUTCString();
    var newcookie = name + "=" + val + ";" + expires + ";path=" + document.location.pathname;
    document.cookie = newcookie;
    console.log("Set:", newcookie)
  },
  UpdateSettings: function(settings) {
    TableRenderer.UpdateSettings(settings);
  },
  GetPlayerId: function() {
    var m = document.cookie.match(/playerid=(\w+)/)
    if (m) {
      return m[1];
    }
  },
  GetPlayerKey: function() {
    var m = document.cookie.match(/playerkey=(\w+)/)
    if (m) {
      return m[1];
    }
  },
  LAST_STATE: "NOSTATE",
  LAST_DATA: {},
  UpdateState: function(doc) {
    var isNew = false;
    if (Poker.LAST_STATE != doc.State) {
      isNew = true;
      Poker.LAST_STATE = doc.State;
    }
    Poker.LAST_DATA = doc;
    var handler = TableRenderer;
    if (doc.State == "NOGAME") {
      // Listing of players currently registered, and ability to register.
      handler = Signup;
    }
    if (isNew) {
      handler.Start();
    }
    handler.Update(doc);
  },
  ProcessEvent: function(evt) {
  },
  Update: function(resp) {
    Poker.UpdateState(resp);
    if (resp.Events) {
      for (var i = 0; i < resp.Events.length; i++) {
        var evt = resp.Events[i];
        if (Events[evt.Event]) {
          Events[evt.Event](evt);
        } else {
          console.log("Don't know what to do with ", evt.Event);
        }
      }
    }
  },
  SendCommand: function(command, args) {
    var params = {
      Name: Poker.NAME,
      PlayerId: Poker.PLAYER_ID,
      PlayerKey: Poker.PLAYER_KEY,
      Command: command,
      Args: args,
    };
    $.ajax({
      url: '/Poker',
      type: 'POST',
      dataType: 'json',
      data: JSON.stringify(params),
      success: function(result) {
        console.log(result);
      }
    });
  },
  Monitor: function() {
    // Start monitoring the state document.
    if (Poker.NAME && Poker.NAME != "") {
      var db = firebase.firestore();
      Poker.DOCUMENT = db.doc("/public/" + Poker.NAME);
      Poker.DOCUMENT.onSnapshot(function(doc) {
        Poker.Update(doc.data());
      });
    }
  },
  InitJqueryTouch: function() {
    function touchHandler(event) {
    var touch = event.changedTouches[0];

    var simulatedEvent = document.createEvent("MouseEvent");
    simulatedEvent.initMouseEvent({
        touchstart: "mousedown",
        touchmove: "mousemove",
        touchend: "mouseup"
    }[event.type], true, true, window, 1,
        touch.screenX, touch.screenY,
        touch.clientX, touch.clientY, false,
        false, false, false, 0, null);

    touch.target.dispatchEvent(simulatedEvent);
    event.preventDefault();
    }

    document.addEventListener("touchstart", touchHandler, true);
    document.addEventListener("touchmove", touchHandler, true);
    document.addEventListener("touchend", touchHandler, true);
    document.addEventListener("touchcancel", touchHandler, true);
  },
};

$(document).ready(function() {
  Poker.InitJqueryTouch();
  Poker.SetupFirebase();
  Poker.Setup();
  Poker.Monitor();
});
