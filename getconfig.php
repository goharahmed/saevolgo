<?php

$db_connection = pg_connect("host=DBSERVERIP port=10432 dbname=kamailio user=kamailio password=kama1l1o");


$queue_name_joined=$_REQUEST["CC-Queue"];
$event_function=$_REQUEST["Event-Calling-Function"];
//list($clientid,$queue_name)=split('_',$queue_name_joined);


$query="SELECT name,strategy,moh_sound,time_base_score,tier_rules_apply,tier_rule_wait_second,tier_rule_wait_multiply_level,tier_rule_no_agent_no_wait,discard_abandoned_after,abandoned_resume_allowed,max_wait_time,max_wait_time_with_no_agent,max_wait_time_with_no_agent_time_reached,record_template from queues where name='".$queue_name_joined."';";

$result = pg_query($db_connection,$query);
$resultArr = pg_fetch_assoc($result);

if (!$resultArr) {
// No Queue found with this Name - return Error to FreeSwitch
        header('Content-Type: text/xml');
        printf("<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"no\"?>\n");
        printf("<document type=\"freeswitch/xml\">\n");
        printf("<section name=\"result\">");
        printf("<result status=\"not found\" />");
        printf("</section>");
        printf("</document>\n");

}else{
        $response = <<< XML
<?xml version="1.0" encoding="UTF-8" standalone="no"?>
        <document type="freeswitch/xml">
                <section name="configuration">
                        <settings>
                        </settings>
                        <configuration name="callcenter.conf" description="CallCenter">
                                <queues>
                                        <queue name="$queue_name_joined"/>
                                        <param name="strategy" value="$resultArr[strategy]"/>
                                        <param name="moh-sound" value="$resultArr[moh_sound]"/>
                                        <param name="time-base-score" value="$resultArr[time_base_score]"/>
                                        <param name="max-wait-time" value="$resultArr[max_wait_time]"/>
                                        <param name="max-wait-time-with-no-agent" value="$resultArr[max_wait_time_with_no_agent]"/>
                                        <param name="max-wait-time-with-no-agent-time-reached" value="$resultArr[max_wait_time_with_no_agent_time_reached]"/>
                                        <param name="tier-rules-apply" value="$resultArr[tier_rules_apply]"/>
                                        <param name="tier-rule-wait-second" value="$resultArr[tier_rule_wait_second]"/>
                                        <param name="tier-rule-wait-multiply-level" value="$resultArr[tier_rule_wait_multiply_level]"/>
                                        <param name="tier-rule-no-agent-no-wait" value="$resultArr[tier_rule_no_agent_no_wait]"/>
                                        <param name="discard-abandoned-after" value="$resultArr[discard_abandoned_after]"/>
                                        <param name="abandoned-resume-allowed" value="$resultArr[abandoned_resume_allowed]"/>
                                </queues>

                        </configuration>
                </section>
        </document>
XML;
die($response);

}
?>
