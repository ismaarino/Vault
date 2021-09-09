let selected = "";
let idk = "<li><a class='dropdown-item disabled' href='#'>Nothing 🤷‍♂️</a></li>";
let vault = {
    get: function(k) {
            astilectron.sendMessage("get~"+k, function(message) {
                if(message != ""){
                    selected = k;
                    $("#key-label").html(message);
                    $("#show-k-layer").fadeIn("fast");
                    $("#result-list").html(idk);
                }
            });
    },
    delete: function() {
        if(selected != ""){
            astilectron.sendMessage("del~"+selected, function(message) {
                $("#key-label").html("");
                $("#show-k-layer").fadeOut("fast");
            });
            selected = "";
            $("#result-list").html(idk);
        }
    },
    search: function(v) {
        if(v != ""){
            astilectron.sendMessage("search~"+v, function(message) {
                if(message != ""){
                    var content = ""
                    message.split("~").forEach(e => content+="<li onclick=\"vault.get('"+e+"')\"><a class='dropdown-item' href='#'>"+e+"</a></li>");
                    $("#result-list").html(content);
                } else {
                    $("#result-list").html(idk);
                }
            });
        } else {
            $("#result-list").html(idk);
        }
    },
    add: function() {
        if(document.getElementById("add-pass").value == document.getElementById("add-pass2").value){
            astilectron.sendMessage("add~"+document.getElementById("add-site").value+"~"+document.getElementById("add-user").value+"~"+document.getElementById("add-pass").value, function(message) {
                console.log("received " + message)
            });
            document.getElementById("add-pass").value = "";
            document.getElementById("add-pass2").value = "";
            document.getElementById("add-user").value = "";
            document.getElementById("add-site").value = "";
        }
    },
    sendKey: function() {
        astilectron.sendMessage("key~"+document.getElementById("key").value, function(message) {
            console.log("received " + message)
        });
        document.getElementById("key").value = "";
    },
    init: function() {
        asticode.loader.init();
        asticode.modaler.init();
        asticode.notifier.init();
    }

};