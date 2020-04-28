// VIEWS contains the renderers
var VIEWS = {};

var Game = {
  name: "", // Game.name is table/{name}/...
  data: {}, // The full game structure.
  watchers: {}, // Firestore watchers / onSnapshots.
};

var Player = {
  uid: "",   // UID, set from firebase auth
  state: "", // Player[UID].State from game structure
  info: "",  // All info from player[uid].state
};

var RollPoker = {
  HEADERS: {},
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
  Auth: function(cb) {
    firebase.auth().onAuthStateChanged(function(user) {
      if (user) {
        console.log("Registered as " + user.displayName);
        Player.uid = user.uid;
        Player.authuser = user;
        user.getIdToken().then(function(token) {
          RollPoker.HEADERS["Authorization"] = "Bearer " + token;
        });
        cb();
      } else {
        console.log("Not registered");
        RollPoker.Update({
          RoomState: "Register",
          Players: [],
        });
      }
    });
  },
  Setup: function() {
    // First make sure we have our game name.
    var m = document.location.pathname.match(/table\/(\w+)$/);
    if (m) {
      Game.name = m[1];
    } else {
      return;
    }

    // Start monitoring.
    RollPoker.Monitor();
    RollPoker.HandleEvents();
  },
  HandleEvents: function() {
    function bind(ename, name) {
      $(window).on(ename, function(evt) {
        if (RollPoker.Handler._handleEvent) {
          RollPoker.Handler._handleEvent(name, evt);
        }
      });
    }
    var eventmap = {
      "click touchstart": "Click",
      "submit": "Submit",
    };
    for (var ename in eventmap) {
      bind(ename, eventmap[ename]);
    }
  },
  Update: function(doc) {
    Player.state = undefined;
    _.each(doc.Players, function(p) {
      if (p.PlayerId == Player.uid) {
        Player.info = p;
      }
    });
    Game.data = doc;
    if (Game.LAST_STATE != doc.RoomState) {
      Game.LAST_STATE = doc.RoomState;
      RollPoker.Handler = VIEWS[doc.RoomState];
      if (RollPoker.Handler) {
        RollPoker.Handler.Start();
      }
    }
    if (RollPoker.Handler) {
      RollPoker.Handler.Update(doc);
    }
  },
  SendCommand: function(command, args) {
    var params = {
      Name: Game.name,
      Command: command,
      Args: args,
    };
    var headers = {};
    $.ajax({
      url: '/Poker',
      type: 'POST',
      dataType: 'json',
      headers: RollPoker.HEADERS,
      data: JSON.stringify(params),
      success: function(result) {
        // TODO: Errors will start returning strings.
        console.log(result);
      }
    });
  },
  Monitor: function() {
    // Start monitoring the state documents.
    var db = firebase.firestore();
    var docref = db.doc("/games/" + Game.name);
    Game.watchers.data = docref.onSnapshot(function(doc) {
      RollPoker.Update(doc.data());
    }, function(error) {
      RollPoker.Update({
        RoomState: "Signup",
        Players: [],
      });
    });

    var logref = db.collection("/games/" + Game.name + "/log")
    logref.orderBy("Timestamp", "desc").limit(30).get().then(function(logs) {
      RollPoker.ProcessLogs(logs, false);
      // Then start a tail.
      Game.watchers.logs = logref.orderBy("Timestamp", "desc").limit(1).onSnapshot(function(logs) {
        RollPoker.ProcessLogs(logs, true);
      });
    });
  },
  LATEST_SEEN: 0,
  LogCallback: undefined,
  UpdateLog: function(log) {
    if (RollPoker.LogCallback) {
      RollPoker.LogCallback(log.Message);
    }
  },
  ProcessLogs: function(logs, doevents) {
    // We get them in an ordered descent. Reverse 'em.
    var rev = [];
    logs.forEach(function(log) {
      rev.push(log.data());
    });
    for (var i = rev.length - 1; i >= 0; i--) {
      var litems = rev[i];
      if (RollPoker.LATEST_SEEN < litems.Timestamp) {
        if (litems.Logs) {
          RollPoker.LATEST_SEEN = litems.Timestamp;
          for (var j = 0; j < litems.Logs.length; j++) {
            var litem = litems.Logs[j];
            if (litem.Message && litem.Message != "") {
              RollPoker.UpdateLog(litem);
            } else if (doevents) {
              var evt = Events[litem.EventName];
              if (evt) {
                evt.apply(evt, litem.Args);
              } else {
                console.log("No Events[" + litem.EventName + "]!");
              }
            }
          }
        }
      }
    }
  },
};

$(document).ready(function() {
  RollPoker.SetupFirebase();
  RollPoker.Auth(function() {
    RollPoker.Setup();
  });
});
