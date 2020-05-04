VIEWS.Register = new View({
  Templates: {
    View: "#registerview",
  },
  Start: function() {
    this.init();
    this.UI = new firebaseui.auth.AuthUI(firebase.auth());
    $('#sizer').html(this.T.View({data: {Players: []}}));
    this.UI.start("#firebase-register", {
      signInOptions: [
        {
          provider: firebase.auth.EmailAuthProvider.PROVIDER_ID,
          signInMethod: firebase.auth.EmailAuthProvider.EMAIL_LINK_SIGN_IN_METHOD
        },
        firebase.auth.GoogleAuthProvider.PROVIDER_ID,
      ],
    });
  },
  Update: function(state) {
    console.log("Register should never get Updates");
  },
});
