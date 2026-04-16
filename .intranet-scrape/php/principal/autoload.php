<?php  /* 2009-07-25 - Autoload Classes */

//if (function_exists('imap_open')) {
function histrix_autoload($class_name) {
    $dir = '';
    $pos  = strpos($class_name, '_');

    /*
    // get current Dir
    if (is_dir('classes')){
        $basepath = '/';
    }
    else {
        $basepath = '/../';
    }
    */
    $basepath = '/../';
    
    if ($pos > 0) {
        $dir = substr($class_name, 0, $pos).'/';
        //    echo $dir.$class_name.'<br>';
    } else {
        if (is_dir(dirname(__FILE__) . $basepath.'classes/'. $class_name)){
            $dir = $class_name.'/';
        }
    }
    $file = dirname(__FILE__) . $basepath.'classes/' .$dir. $class_name . '.php';
    
    if (is_file($file)) {
        require($file);
    }
    else {
        //throw new Exception("Unable to load $file.");
        // search within xml foldes
        $datapath= $_SESSION['datapath'];
        $dirname = '../database/'.$datapath.'xml';
//	die($class_name);
	//loger($class_name, 'class_search.log');
        // just class name
        $files= glob_recursive($dirname.'/*/'.$class_name.'.php');        
        $file = $files[0] ;
        if (is_file($file)) {
            require($file);
        }
        else {
        // with 'class' appended 
            $files= glob_recursive($dirname.'/*/'.$class_name.'.class.php');        
            $file = $files[0] ;
        
            if (is_file($file)) {
                require($file);
            }
        }

    }
    
}
//}
$var = spl_autoload_register('histrix_autoload', false);


function glob_recursive($pattern, $flags = 0)
    {
        $files = glob($pattern, $flags);
       
        foreach (glob(dirname($pattern).'/*', GLOB_ONLYDIR|GLOB_NOSORT) as $dir)
        {
            $files = array_merge($files, glob_recursive($dir.'/'.basename($pattern), $flags));
        }
       
        return $files;
    }
?>
