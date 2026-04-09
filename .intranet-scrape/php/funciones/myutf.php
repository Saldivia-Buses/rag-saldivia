#!/usr/bin/php
<?php

function q($text, $c) {
    if (mysql_query($text, $c) === FALSE) {
            echo('failed at: ' . $text . "\n\n" . mysql_error($c) . "\n");
    }
}
                
function c($hostname, $username, $password, $database, $charset, $collation) {

    $c = mysql_connect($hostname, $username, $password) ;
    
    if(!$c) die('Could not connect: ' . mysql_error() . "\n");
    
    

    mysql_select_db($database, $c);
    $qTable = mysql_query("SHOW FULL TABLES WHERE Table_Type = 'BASE TABLE' ", $c);

    while($table = mysql_fetch_row($qTable)) 
    {
        $table = $table[0];
        echo "Working on $table ... \n";
        $qColumn = mysql_query('SHOW FULL COLUMNS FROM `' . $table . '`', $c);

        while($column = mysql_fetch_assoc($qColumn)) 
        {
            if($column['Type'] != 'date' && $column['Type'] != 'timestamp') 
            {
                q('ALTER TABLE `' . $table . '` CHANGE `' . $column['Field'] . '` `' . $column['Field'] . '` ' . $column['Type'] . ($column['Collation'] === null ? ' ' : ' CHARACTER SET ' . $charset . ' COLLATE ' . $collation . ' ') . ($column['Null'] === 'YES' ? 'NULL' : 'NOT NULL') . ($column['Default'] === null ? '' : ' DEFAULT \'' . $column['Default'] . '\'' ) . ' ' . strtoupper($column['Extra']), $c);
	    }
        }

        q('ALTER TABLE `' . $table . '` DEFAULT CHARACTER SET ' . $charset . ' COLLATE ' . $collation, $c);
    }

    q('ALTER DATABASE `' . $database . '` DEFAULT CHARACTER SET ' . $charset . ' COLLATE ' . $collation, $c);

    mysql_close($c);
}

if(defined('STDIN') ) {                                                                                        

    if( isset($argv[1]) && isset($argv[2]) )
    {
	c('localhost', 'root', "{$argv[2]}", "{$argv[1]}", 'utf8', 'utf8_general_ci');
        echo "finished without failure\n";
    } 
    else
	die("Bad Parameters\n");
}
else
    die("Must be shell executed\n");
                                                                                                            