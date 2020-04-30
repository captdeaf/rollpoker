var App = {
  Setup: function() {
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

    firebase.auth().onAuthStateChanged(function(user) {
      if (user) {
        $("#register").hide();
        $("#welcome").text("Logged in as:" + user.displayName).show();
        $("#makeroom").show();
        $("#gamesettings").on("submit", function(evt) {
          console.log("Creating game");
          evt.stopPropagation();
          evt.preventDefault();
          App.CreateGame(user);
        });
      } else {
        // Are we logged in?
        $("#register").on("click touchstart", function() {
          var ui = new firebaseui.auth.AuthUI(firebase.auth());
          ui.start('#registerui', {
            signInOptions: [
              {
                provider: firebase.auth.EmailAuthProvider.PROVIDER_ID,
                signInMethod: firebase.auth.EmailAuthProvider.EMAIL_LINK_SIGN_IN_METHOD
              },
              firebase.auth.GoogleAuthProvider.PROVIDER_ID,
            ],
          });

        });
      }
    });
  },
  CreateGame: function(user) {
    console.log("getIdToken then");
    user.getIdToken().then(function(token) {
      console.log("got IdToken");
      var data = {}
      _.each($('#gamesettings').serializeArray(), function(fd) {
        data[fd.name] = fd.value;
      });
      data["DisplayName"] = user.displayName;
      console.log(data);
      $.ajax({
        url: '/MakeTable',
        type: 'POST',
        dataType: 'json',
        headers: {Authorization: "Bearer " + token},
        data: JSON.stringify(data),
        success: function(result) {
          console.log(result);
          document.location = "/table/" + result.Name;
        },
        error: function(xhr, resp, text) {
          console.log(xhr, resp, text);
        },
      })
    });
  },
};


$(document).ready(function() {
  App.Setup();
});
