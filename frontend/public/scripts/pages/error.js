export { ErrorP }


const ErrorP = {
 
    html: ` 
    <head>
     <link rel="stylesheet" href="/public/styles/error.css">
    </head> 
   <div id="clouds">
            <div class="cloud x1"></div>
            <div class="cloud x1_5"></div>
            <div class="cloud x2"></div>
            <div class="cloud x3"></div>
            <div class="cloud x4"></div>
            <div class="cloud x5"></div>
        </div>
        <div class='main'>
            <div id='statuscode'></div>
            <hr>
            <div id='statusmsg'></div>
            <a class='btn' href='/' data-link>GO BACK</a>
        </div>
  `,
  setup: (code, msg) => {
    document.getElementById('statuscode').textContent = code 
    document.getElementById('statusmsg').textContent = msg

  }
};

// TODO enhance error page