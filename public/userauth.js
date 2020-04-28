$(document).ready(function() {
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


});
