/* 
 * jQuery Histrix plugin for drag and drop files
 */


(function( $ ) {

    var count = 0;
    var k = 1024;
    var m = k * k;
    var maxFileSizeMB = 25;
    var width = 200;
    var partialCount = 0;
    var completed = 0;
    var targetObject = '';
    var lastFileName = '';
    var settings = {
        'uploader'         : 'uploadhtml5.php',
        'destinationDir'   : '/',
        'maxWidth'   	   : 0,
        'output'           : this,
        'speedStatus'      : '#speedStatus'
    };

    $.fn.htxfiledrop = function(options) {

        if ( options ) {
            $.extend( settings, options );
        }

        this.live('drop', dodrop)
        .live('dragenter', dragEnter)
        .live('dragover', dragOver)
        .live('dragleave', dragExit);

    }

    function updateBytes(evt) {
        if (evt.lengthComputable) {
            evt.target.curLoad = evt.loaded;
            evt.target.log.parentNode.parentNode.previousSibling.textContent =
            Number(evt.loaded/k).toFixed() + "/"+ Number(evt.total/k).toFixed() + "kB";
        }
    }

    function updateSpeed(target) {
        if (!target.curLoad) return;
        if (target.log.parentNode != null){
            target.log.parentNode.parentNode.previousSibling.previousSibling.textContent =
            Number((target.curLoad - target.prevLoad)/k).toFixed() + "kB/s";
        }
        target.prevLoad = target.curLoad;
    }

    function updateProgress(evt) {
        updateBytes(evt);
        if (evt.lengthComputable) {
            var loaded = (evt.loaded / evt.total);
            if (loaded < 1) {
                var newW = loaded * width;
                if (newW < 10) newW = 10;
                evt.target.log.style.width = newW + "px";
            }
        }
    }

    function loadError(evt) {
        evt.target.log.setAttribute("status", "error");
        evt.target.log.parentNode.parentNode.previousSibling.previousSibling.textContent = "error";
        clearTarget(evt.target);
    }

    function loaded(evt) {
        updateBytes(evt);
        evt.target.log.style.width = width + "px";
        evt.target.log.setAttribute("status", "loaded");
        evt.target.log.parentNode.parentNode.previousSibling.previousSibling.textContent = "";
        completed += 1;
        doRefresh();
    }

    function clearTarget(target) {
        clearInterval(target.interval);
        target.onprogress = null;
        target.onload = null;
        target.onerror = null;
    }

    function initXHREventTarget(target, container) {
        var speed = document.createElement("td");
        speed.className = "speed";
        container.appendChild(speed);

        var info = document.createElement("td");
        info.className = "info";
        container.appendChild(info);

        var progressContainerTd = document.createElement("td");
        container.appendChild(progressContainerTd);

        var progressContainer = document.createElement("div");
        progressContainer.className = "progressBarContainer";
        progressContainerTd.appendChild(progressContainer);

        var progress = document.createElement("div");
        progressContainer.appendChild(progress);
        progress.className = "progressBar";

        target.log = progress;
        target.interval = setInterval(updateSpeed, 1000, target);
        target.curLoad = 0;
        target.prevLoad = 0;
        target.onprogress = updateProgress;
        target.onload = loaded;
        target.onerror = loadError;
    }


    function start(file, item) {

        var xhr = new XMLHttpRequest();
        ++count;

        var container = document.createElement("tr");

        var filename = document.createElement("td");
        container.appendChild(filename);


        var fileName  = file.fileName || file.name;
        var fileSize  = file.fileSize || file.size;

        lastFileName = fileName;

        filename.textContent = fileName;
        filename.className = "filename";

        initXHREventTarget(xhr.upload, container, item);

	$(settings.speedStatus).css({display:''});
        var tbody = $(settings.speedStatus)[0];
        tbody.appendChild(container);
        tbody.style.display = "";

                
        xhr.open("POST", settings.uploader+"?dir="+ settings.destinationDir+'&maxWidth='+settings.maxWidth , true);

        xhr.setRequestHeader("Cache-Control", "no-cache");
        xhr.setRequestHeader("X-Requested-With", "XMLHttpRequest");
        xhr.setRequestHeader("X-File-Name", encodeURIComponent(fileName));
        xhr.setRequestHeader("X-File-Size", fileSize);

        var boundary = "AJAX--------------" + (new Date).getTime();
        var contentType = "multipart/mixed; boundary=" + boundary;
        xhr.setRequestHeader("Content-Type", contentType);

        var send = false;
        try {
            xhr.sendAsBinary(file.getAsBinary());      //firefox method
            send = true;
        }
        catch(ex){
            send = false;
        }

        if (!send)
            xhr.send(file); //chrome support
    }


    function dodrop(event){

        event.stopPropagation();
        event.preventDefault();


        var dest = $(this).closest('[destinationDir]');

        settings.destinationDir = dest.attr('destinationDir') || settings.destinationDir;
    	settings.maxWidth 	= dest.attr('maxWidth') || settings.maxWidth;


        var pathobj = $('#'+ dest.attr('pathObj')).val();
        if (pathobj)
            settings.destinationDir += pathobj;

        settings.destinationDir +=  '/';

        var event2 = event.originalEvent;
        var dt = event2.dataTransfer;

        var files = dt.files;

        var count = files.length;
        partialCount = count;

        output("Files: " + count + "\n");

        $(settings.speedStatus).css({
            display:'block'
        });

        for (var i = 0; i < count; i++) {
            try {
                var file = files.item(i);
                start(file, i);
            } catch (ex) {
               
                output("<<error>>\n");
            }
        }

    }

    function doRefresh(){

        if (completed == partialCount){

            $(targetObject).removeClass('drop');
            $(settings.speedStatus).fadeOut('slow', function(){
                $(settings.speedStatus).empty();
            });

            //var dest = $(targetObject).closest('[destinationDir]');
            $('#'+ $(targetObject).attr('fileinput')).val(lastFileName).change();

            $('#RefreshFileManager').click();
            completed = 0;
            partialCount = 0;
            count = 0;
            
        }
        else loger('Completos:'+completed);
    }

    function output(text){
       
        $(settings.output).append(text);
    }

    function dragEnter(event) {
        var target = event.target;

        targetObject = $(target).closest('[destinationDir]');
        $(targetObject).addClass('drop');
    }

    function dragExit(event) {
        $(targetObject).removeClass('drop');
        event.stopPropagation();
        event.preventDefault();
    }
    function dragOver(event) {

        event.stopPropagation();
        event.preventDefault();
    }

})( jQuery );