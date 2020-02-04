<?php
/**
 * Global Configuration Override
 *
 * You can use this file for overriding configuration values from modules, etc.
 * You would place values in here that are agnostic to the environment and not
 * sensitive to security.
 *
 * @NOTE: In practice, this file will typically be INCLUDED in your source
 * control, so do not include passwords or other sensitive information in this
 * file.
 */

$vcapServices = \Zend\Json\Json::decode(getenv('VCAP_SERVICES'), \Zend\Json\Json::TYPE_ARRAY);
$clearDbCreds = $vcapServices['cleardb'][0]['credentials'];

return array(
    'db' => array(
        'driver'    => 'PdoMysql',
        'hostname'  => $clearDbCreds['hostname'],
        'database'  => $clearDbCreds['name'],
        'username'  => $clearDbCreds['username'],
        'password'  => $clearDbCreds['password'],
    ),
    'scn-social-auth' => array(
        'facebook_client_id' => getenv('facebook_client_id'),
        'facebook_secret' => getenv('facebook_secret'),
        'twitter_consumer_key' => getenv('twitter_consumer_key'),
        'twitter_consumer_secret' => getenv('twitter_consumer_secret'),
    ),
    'service_manager' => array(
        'factories' => array(
            'Zend\Db\Adapter\Adapter' => 'Zend\Db\Adapter\AdapterServiceFactory',
        ),
        'invokables' => array(
            'Zend\Session\SessionManager' => 'Zend\Session\SessionManager',
        ),
    ),
);
