<?php 
/* DEBUG SQL QUERYS */
include ("./autoload.php");
include ("./sessionCheck.php");
include ("../lib/sqlSintax/SqlHighlighter.class.php");

if ($_SESSION['EDITOR']== 'editor'){
    $xml = '_'.$_GET['xml'];

    $instance = $_REQUEST['instance'];
    if ($instance != ''){
        $MisDatos = new ContDatos("");
        $MisDatos = Histrix_XmlReader::unserializeContainer(null, $_REQUEST['instance']);
        
        

        $mySQLhighlighted = new SqlHighlighter();

        echo '<div >';
        echo '<h1>Last Select</h1>';

   //     echo $mySQLhighlighted->highlight( $update );
        
        echo $mySQLhighlighted->highlight( $MisDatos->_lastSelect );
        echo '</div>';        
    }
    /*
    if ($xml != ''){

        $update = print_r($_SESSION['last_update_sql'][$xml], true);
        $select = print_r($_SESSION['last_select_sql'][$xml], true);
        $mySQLhighlighted = new SqlHighlighter();

        echo '<div >';
       echo '<h1>Select</h1>';

        echo $mySQLhighlighted->highlight( $update );
        
        echo $mySQLhighlighted->highlight( $select );
        echo '</div>';
    }

    */  
        if (function_exists('memcache_connect')) {
            $memcache = new Memcache();
            $connection = @$memcache->connect('localhost',11211);
            
            if ($connection){
                echo 'Memcache: Habilitado';
            }
            else echo 'Memcache: Deshabilitado';
        }
        echo '<hr>';
    
       echo '<h1>Sesion</h1>';
       if (function_exists('igbinary_serialize')){
            echo '<p>Relative size (igbinary): '.strlen(igbinary_serialize($_SESSION)).'</p>';
       } else {
            echo '<p>Relative size  (php): '.strlen(serialize($_SESSION)).'</p>';
       }
       

       echo '<p>Open xml files: '.count($_SESSION['xml']).'</p>';
        echo '<pre>';
        print_r($_SESSION);
        echo '</pre>';
        echo '<hr>';

    
}
else
    echo 'NO DEBUG ALLOWED';
?>