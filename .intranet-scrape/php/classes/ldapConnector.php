<?php
/* 
 * Ldap connector 2009/05/20
 * Generate Ldap registry, CRUD operations
 * 
 */

class ldapConnector{


    function __construct($server='', $bindDn ='', $rdn='', $pass=''){
        $server = 'localhost';
        $bindDn = 'cn=admin,dc=saldivia,dc=com,dc=ar';
        $rdn    = 'dc=saldivia,dc=com,dc=ar';
        $pass   = 'm2450e';

        $this->server = $server;
        $this->bindDn    = $bindDn;
        $this->rdn    = $rdn;
        $this->pass   = $pass;
        $this->connect($server, $bindDn, $pass);
    }

    function addData($key, $value){
        if (trim($value) != '')
            $this->Data[$key] = $value;

    }

    /**
     * add Old Key Pair Value (for updates /deletes)
     * @param string $key
     * @param string $value
     */
    function addKeyOld($key, $value){
        $this->KeyOld[$key] = $value;
    }

    /**
     * add New Key Pair Value (for Insertes)
     * @param string $key
     * @param string $value 
     */
    function addKeyNew($key, $value){
            $this->KeyNew[$key] = $value;
    }


    /**
     * insert Ldap record
     * @param string $dn unique identifier
     */
    function insert($dn=''){
            // add data to directory
        if ($dn == ''){
            foreach($this->KeyNew as $keyName => $keyValue){
                if (trim($keyValue) == '') return false;
                $dn .= $keyName.'='.$keyValue.',';
            }
        }
        $this->Data['objectClass'][0] = "top";
        $this->Data['objectClass'][1] ="inetOrgPerson";
        $this->Data['objectClass'][2] ="mozillaAbPersonAlpha";

        $key = $dn.$this->rdn;
    //    print_r($this->Data);
        $r = @ldap_add($this->ds, $key, $this->Data);
    }


    /**
     * @param string $dn
     * @return boolean
     */
    function update($dn=''){

        // generate new dn
        if ($dn == ''){
	
	    // Sanitice
	
            foreach($this->KeyOld as $keyName => $keyValue){
                if (trim($keyValue) == '') {
                    // if old key value was blank INSERT new record
                    $this->insert(); 
                    return false;
                }
                $dn .= $keyName.'='.$keyValue.',';
            }
            foreach($this->KeyNew as $keyName => $keyValue){
                $dn2 .= $keyName.'='.$keyValue.',';
            }
                // if key changes
            if ($dn != $dn2){
                $this->delete();
                $this->insert();
                return true;
            }

        }

        $this->Data['objectClass'][0] = "top";
        $this->Data['objectClass'][1] ="inetOrgPerson";
        $this->Data['objectClass'][2] ="mozillaAbPersonAlpha";

        $key =  $dn.$this->rdn;
/*	print_r($this->ds);
        print_r($key);
	print_r($this->Data);
	*/
        $r = @ldap_modify($this->ds, $key, $this->Data);
	if (!$r) $this->insert();
        return true;
    }

    /**
     * Delete ldap record
     * @param string $dn
     * @return <type>
     */
    function delete($dn=''){
        //delete entry
        if ($dn == ''){
            foreach($this->KeyOld as $keyName => $keyValue){
                if (trim($keyValue) == '') return false;
                $dn .= $keyName.'='.$keyValue.',';
            }
        }
        $key = $dn.$this->rdn;
        $del = @ldap_delete( $this->ds  , $key  );
    }

    /**
     * Connect to ldap Server
     * @param string $server
     * @param string $bindDn
     * @param string $pass
     * @return bind resource
     */
    function connect($server = '', $bindDn='', $pass=''){
         $server = ($server != '')?$server:$this->server;
         $bindDn    = ($bindDn != '')?$bindDn:$this->bindDn;
         $pass   = ($pass != '')?$server:$this->pass;
         if (function_exists('ldap_connect')){
             $ds = ldap_connect($server, 389);
             ldap_set_option($ds, LDAP_OPT_PROTOCOL_VERSION, 3);
             $r = ldap_bind($ds, $bindDn, $pass);

             $this->ds = $ds;
             return $r;
        }
    }

    /**
     * Close ldap connection
     */
    function close(){
        ldap_close($this->ds);
    }

}

?>