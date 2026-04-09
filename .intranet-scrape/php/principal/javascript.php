<?php
$jsPath = '../javascript/';
$javascripts[] = $jsPath.'dtree.js';
$javascripts[] = $jsPath.'sorttable.js';

$javascripts[] = $jsPath.'swfobject.js';
$javascripts[] = $jsPath.'XulMenu.js';


$javascripts[] = $jsPath.'jquery-1.7.2.min.js';
$javascripts[] = $jsPath.'jquery-ui-1.8.16.custom.min.js';
$javascripts[] = $jsPath.'jquery.requireScript-1.2.1.js';

$javascripts[] = $jsPath.'jquery.histrix.filedrop.js'; // custom histrix jquery plugin for drag and drop files

$javascripts[] = $jsPath.'lang/jquery.ui.datepicker-'.$datosbase->lang.'.js';

$javascripts[] = $jsPath.'jquery.touch.compact.js'; // IPAD SUPPORT

$javascripts[] = $jsPath.'jquery.idletimer.js';  // manage timeout

//$javascripts[] = $jsPath.'fullcalendar/fullcalendar.js';			// dynamic load now
//$javascripts[] = $jsPath.'fullcalendar/gcal.js';				// dynamic load now

                                        /*
$javascripts[] = $jsPath.'recurrenceinput/jquery.tmpl-beta1.js';		// dynamic load now
$javascripts[] = $jsPath.'recurrenceinput/jquery.tools.overlay-1.2.7.js';	// dynamic load now
$javascripts[] = $jsPath.'recurrenceinput/jquery.tools.dateinput-1.2.7.js';   // TODO: REFACTOR TO USER JQUERY DATEPICKER
$javascripts[] = $jsPath.'recurrenceinput/jquery.jquery.utils.i18n-0.8.5.js';	// dynamic load now
$javascripts[] = $jsPath.'recurrenceinput/jquery.recurrenceinput.js';		// dynamic load now
$javascripts[] = $jsPath.'recurrenceinput/jquery.recurrenceinput-'.$datosbase->lang.'.js'; 	// dynamic load now
                                             */
//$javascripts[] = $jsPath.'jquery.orgchart.min.js';
$javascripts[] = $jsPath.'jquery.jOrgChart.js';


$javascripts[] = $jsPath.'jquery.lightbox-0.5.min.js';
$javascripts[] = $jsPath.'jquery.json-1.3.min.js';
$javascripts[] = $jsPath.'jquery.scrollTo-min.js';
$javascripts[] = $jsPath.'jquery.autocomplete.min.js';                          // DUPLICATED WITH JQUERY UI
$javascripts[] = $jsPath.'jquery.pulse.js';                                     // DUPLICATED WITH JQUERY UI
$javascripts[] = $jsPath.'JQuerySpinBtn.js';
$javascripts[] = $jsPath.'jquery.calculation.min.js';  // Form Calculation


$javascripts[] = $jsPath.'jquery.tablednd_0_5.js';  // Drag And Drop

// $javascripts[] = $jsPath.'jquery.tinyTips.js';      // jquery tooltips // dynamic load now
// $javascripts[] = $jsPath.'jwysiwyg/jquery.wysiwyg.js'; // dynamic load now
// $javascripts[] = $jsPath.'colorpicker.js'; 		  // dynamic load now

$javascripts[] = $jsPath.'jquery.maskedinput-1.2.2.min.js'; // Mask Input
$javascripts[] = $jsPath.'webforms2/webforms2_src.js';

$plugins = Cache::getCache('Plugins');

// Hook to registered Plgins javascript
$returnedValues = PluginLoader::executePluginHooks('javascriptLoad', $plugins);
if (is_array($returnedValues)){
    foreach($returnedValues as $plugedJavascript){
            foreach($plugedJavascript as $js){
            $javascripts[] = $js;
        }
    }
}

//$javascripts[] = 'overlib.js';  // Library for mouseover events in charts
//$javascripts[] = 'pMap.js';     // ImageMap in charts

$lastjs_size = 0;
foreach($javascripts as $n => $js) {
    if(is_file($js)) {
        $lastjs_size += filesize($js);
    }
}

?>