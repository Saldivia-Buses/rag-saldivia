<?php
//phpinfo();
        putenv("ODBCINI=/etc/odbc.ini");
	putenv("LD_ASSUME_KERNEL=2.4.0");
	//Line below important - stops a strange Merant 6060 error appearing on stdout
//        putenv("LD_LIBRARY_PATH=/usr/dlc/lib:/usr/dlc/odbc/lib");
    	$link2 = odbc_connect("Mysaldivia", 'root', '');


	$result = @odbc_data_source( $link2, SQL_FETCH_FIRST );
	print_r($result);
    while($result)
    {
   if (strtolower($dsn) == strtolower($result['server'])) {
          echo $result["description"] . "<br>\n";
	         break;
		    }
		           else
			          $result = @odbc_data_source( $link2, SQL_FETCH_NEXT );
				  }
				  


?>