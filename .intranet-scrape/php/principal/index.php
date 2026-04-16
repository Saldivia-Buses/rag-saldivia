<?php
/**
 * Histrix Main program
 */
include ('autoload.php');
include ('../funciones/utiles.php');

// Session Start and check
$DirectAccess = true;
include ('./sessionCheck.php');
include ("../funciones/conexion.php");


/*
if I clear session Data i cant have more than 1 window open

$clearSessionData = true;
include ('./delvars.php');
*/
//$_SESSION['mobile'] = 1;

// Load configuration
$db          = $_SESSION["db"];
$nom_empresa = $datosbase->nombre;
$img_fondo   = $datosbase->img_fondo;
$css         = $datosbase->css;
$supportUrl  = $datosbase->supportUrl;

//$datosbase 	 = $_$config->bases[$db];
$xmlPath    = $datosbase->xmlPath;
$gmapKey     = $datosbase->gmapKey;
$datapath    = '../database/'.$xmlPath;

// Check for CUSTOM Favicon
$favicon = $datapath.'/img/'.'favicon.ico';
if (!is_file($favicon)) $favicon = '../img/histrixico.gif';

$background_image = $datapath.'/img/'.$img_fondo;

/* Tomo los datos de cada conexion */
if($datosbase->css != '') {
    $customcss = $datapath.'css/'.$datosbase->css;
}

//$_SESSION['css'] = $css;
Cache::setCache('css', $css);
$LANG_PATH='../lang/';
// Database Connect


$user = new User($_SESSION['usuario']);
if (count($user->preferences) > 0) {
    foreach($user->preferences as $key => $value){
        $_SESSION[$key]=$value;
    }
    $_SESSION['user_preferences'] = $user->preferences;
}

// PLUGIN LOADING
$pluginLoader = new PluginLoader();

$pluginLoader->getAvailablePlugins();
$plugins = $pluginLoader->getRegisteredPlugins();
Cache::setCache('Plugins', $plugins);

//  PUBLIC "-//W3C//DTD XHTML 1.1//EN" "http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd">

?>
<!DOCTYPE html>
<html>
    <head>
        <meta name="tipo_contenido"  content="text/html;" http-equiv="content-type" charset="utf-8" />
        <meta name="apple-mobile-web-app-capable" content="yes" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        
        <link rel="shortcut icon" href="<?php echo $favicon; ?>" />
        <title>[<?php echo $_SESSION['db']; ?>]</title>
        <link rel="stylesheet" type="text/css" href="../funciones/concat.php?type=css" />
            <?php

            $cssFiles[] = '../css/histrix.css';
            $cssFiles[] = '../css/dtree.css';
            $cssFiles[] = '../css/jquery.lightbox-0.5.css';
            $cssFiles[] = '../css/jquery.autocomplete.css';
            $cssFiles[] = '../javascript/webforms/webforms2.css';
            $cssFiles[] = '../css/xulmenu.css';
            $cssFiles[] = '../css/calendar-blue.css'; //calendar
            $cssFiles[] = '../css/smoothness/jquery-ui-1.8.16.custom.css'; //jquery UI
            $cssFiles[] = '../css/JQuerySpinBtn.css'; //Spin Button
            $cssFiles[] = '../javascript/jwysiwyg/jquery.wysiwyg.css';
            $cssFiles[] = '../javascript/fullcalendar/fullcalendar.css';
            $cssFiles[] = '../css/jquery.jOrgChart.css';
                                         

            $cssFiles[] = '../javascript/recurrenceinput/jquery.recurrenceinput.css';
            $cssFiles[] = '../javascript/recurrenceinput/jquery.tools.dateinput.css';
                                           


            // Hook to registered Css plugins Files
            $returnedValues = PluginLoader::executePluginHooks('cssLoad', $plugins);

            if (is_array($returnedValues)){
                foreach($returnedValues as $plugedCss ){
                    foreach($plugedCss as $css){
                        $cssFiles[] = $css;
                    }
                }
            }

            //$_SESSION['css']= $cssFiles;

            Cache::setCache('css', $cssFiles);

            unset($cssFiles);

            $cssFiles[] = '../css/user.css.php';


            if ($_SESSION['mobile'] == 1){
                $cssFiles[] = '../css/mobile.css';
             }

            if (isset($customcss))
                $cssFiles[] = $customcss;


            foreach($cssFiles as $n => $cssF) {
                if(is_file($cssF)) {
                    $css_size = filesize($cssF);
                    echo '<link rel="stylesheet" type="text/css" href="'.$cssF.'?'.$css_size.'.css" />';
                    echo "\n";
                }
            }
            /*
        if ($gmapKey != ''){
                echo '<script src="http://maps.google.com/maps?file=api&v=2&key='.$gmapKey.'" type="text/javascript"></script>';
        }
        */
        
/*        
    <script id="jquery-recurrenceinput-display-tmpl" type="text/x-jquery-tmpl" src="../javascript/recurrenceinput/jquery.recurrenceinput.display.tmpl" ></script>
    <script id="jquery-recurrenceinput-form-tmpl"     type="text/x-jquery-tmpl" src="../javascript/recurrenceinput/jquery.recurrenceinput.form.tmpl" ></script>
  */      
            ?>
            

    </head>
    <?php
    /* si se inicia session valida se carga el menu */

    $closeButton='window.open(\'close_session.php\', \'_self\')';

    if (isset($_GET['prism']))
    if ($_GET['prism']) {
        $closeButton='window.open(\'../index.php?prism=true\', \'_self\')';
    $closeTitle = 'Cerrar Session';

    $closeImg = '<img width="18px" height="18px" title="'.$closeTitle.'" onclick="'.$closeButton.'" style="cursor:pointer; margin:1px 2px;" align="top" src="../img/logout.png" />';

    }

  //  if ($_SESSION["validado"])
  //      $bodyParams = 'onresize="Histrix.resizeAll();" ';

    // MENUBAR ORIENTATION
    // Horizontal

    $orientation = ($_SESSION['mobile'] == 1)?'v':'h';
     $orientation = 'h';
     $menutype = 'menubar';

     if ($_SESSION['mobile'] == 1){
         $menutype = 'button';

     }

    $histrixMenu = new Histrix_Menu($menutype, $orientation);

    $menubarclass = $orientation.'menubar';
    $tabbarclass  = $orientation.'tabbar';
    $superclass   = $orientation.'super';
    $trayclass    = $orientation.'trayclass';


    ?>
    <body <?php echo $bodyParams; ?> >
        <div class="Pagina" <?php echo ($img_fondo != '')?'style="background-image:url(\''.$background_image.'\');"':''; ?>>

            <div class="<?php echo $menubarclass; ?>">
                <?php

                    $Histrix_Tray = new Histrix_Tray($trayclass, $supportUrl);
                    $Histrix_Tray->render();

                    $tipoMen['login'] = $user->login;
                    $perfil = $_SESSION['profile'];

                    $histrixMenu->build("phpmen", 'myMenu', $perfil, '', 'menu1', $tipoMen);
                    $histrixMenu->render(false, 'position:absolute; left:0px;');
                   

                ?>
            </div>

            <div class="<?php echo $tabbarclass; ?>"><ul class="tabs" id="tabs"></ul></div>

			<table id="speedStatus" style="z-index:119999;display:none;"></table>
            <div id="notifications" class="foreground"></div>
            <div  class="<?php echo $superclass; ?>" id="Supracontenido" >
                    <?php
                    PluginLoader::executePluginHooks('desktopInit', $plugins);
                    ?>
                <div id="postBoard"></div>


            </div>
            <div id="waitScreen" ></div>

            <!-- TODO: make bar optional -->
 
            
            <?php
                $utilbar = true;
                if ($utilbar){
                    echo '<div id="utilbar" class="utilbar utilbarColor"  title="'.$i18n['dlbclickExpand'].'">';
                    include "utilbar.php";
                    echo '<div id="widgets" >';
                    // Hook to register sideBar Plugin functionality
//var_dump($plugins);
//die();
                    PluginLoader::executePluginHooks('sideBarInit', $plugins);
                    echo '</div>';
                    echo '</div>';
                }
//die();
        
            ?>
     
            <div id="barraestado">
                <?php
                    PluginLoader::executePluginHooks('statusBarInit', $plugins);
                    
                    if ($_SESSION['emailUser'] != '') {
                        ?>
                        <div style="cursor:pointer;bottom:20px; position:absolute;z-index:100;" id="bubbleimap"  onclick="$(this).slideToggle();" ></div>
                        <div id="webmail" style="overflow:visible; z-index:100;float:left;margin-left:5px" class="panelApp"  ></div>
                        <?php
                        }
                        ?>
                    <div class="msg" style="visibility:hidden;" id="Msg" ></div>
                    <div id="histrixDebug" ></div>
                    <div class="ultimosconect panelApp shadow"  >
                        <div id="msgBar"><span id="_numberusers">Usuarios</span></div>
                        <div id="mensajeria" class="userlist" style="display:none;"> </div>
                    </div>

            </div>
                <div id="userInfo" class="userInfo" style="display:none;"></div>
            <?php

                include('javascript.php');
                Cache::setCache('javascript', $javascripts);

                // Unified Javascript Libraries (less https calls)
                echo '		<script type="text/javascript" src="../funciones/concat.php?type=javascript"></script>';
                unset ($javascripts);


                $javascripts[] = '../javascript/histrix.js';
                $javascripts[] = '../javascript/lang/histrix-'.$datosbase->lang.'.js';   

                foreach($javascripts as $n => $js) {
                    if(is_file($js)) {
                        
                        $js_size = filesize($js);
                        $lastjs_size += $js_size;
                        echo '		<script id="'.$n.'" type="text/javascript" src="'.$js.'?'.$js_size.'"></script>';

                    }
                    else {
                        echo '		<script id="'.$n.'" type="text/javascript" src="'.$js.'"></script>';

                    }
                    
                }
                $_SESSION['lastjs_size'] = $lastjs_size;
                
                ?>
            <br/><br/>
            <div style="display:none;">
                <audio id="audioding"  style="display:none;" controls="false"  width="0" height="0">
                     <source src="../audio/notify.ogg" type="audio/ogg; codecs=vorbis"/>
                     <source src="../audio/notify.mp3" />
                </audio>
                <audio  id="audiochan" style="display:none;"  width="0" height="0">
                     <source src="../audio/error.ogg" type="audio/ogg; codecs=vorbis"/>
                     <source src="../audio/error.mp3" />
                </audio>
                <audio  id="audioselect"style="display:none;"  width="0" height="0">
                     <source src="../audio/select.ogg" type="audio/ogg; codecs=vorbis"/>
                     <source src="../audio/select.mp3" />
                </audio>
                <audio  id="audiochat"style="display:none;"  width="0" height="0">
                     <source src="../audio/chat.ogg" type="audio/ogg; codecs=vorbis"/>
                     <source src="../audio/chat.mp3" />
                </audio>
                
            </div>
            <script type="text/javascript" >
                    $(document).ready(function(){

                            


                            Histrix.lang = '<?php echo $_SESSION['lang']; ?>';
        		            Histrix.db   = '<?php echo $_SESSION['db']; ?>';
        		            Histrix.user = '<?php echo $_SESSION['usuario']; ?>';

        		            Histrix.init(window.opener);
                                                 
                            // Unity integration
			try{
                            if (external.getUnityObject){
                                Histrix.Unity = external.getUnityObject(1.0); 
                                Histrix.Unity.init({name: "<?php echo $nom_empresa; ?>",
                                            iconUrl: "<?php echo $_SERVER['HTTP_REFERER']."img/histrix_button.128.png"; ?>",
                                            onInit: function() {
                                  // Integrate with Unity!


                                    Histrix.Unity.addAction("/About Histrix", function(){
                                        var htxwin =  window.open( 'http://www.estudiogenus.com' );
                                    });
                                   
                                    Histrix.Unity.Launcher.addAction("About Histrix", function(){
                                       var htxwin =  window.open( 'http://www.estudiogenus.com' );
                                    });
                                    Histrix.Unity.Launcher.clearCount();   


                                    Histrix.Unity.Notification.showNotification("Histrix", "Notifications enabled");     
                                    
                                }
                        	});
                            }
			} catch(e){}
                                
                             

		            <?php
		            // Javascript plugin custom Code
		            	$pluginOutput = PluginLoader::executePluginHooks('javascriptInit', $plugins) ;
		            	if ( is_array($pluginOutput)) 
		               		print_r( implode( ' ', $pluginOutput));
		            ?>
		});

            </script>
       </body>
</html>
