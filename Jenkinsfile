pipeline {
    agent any

    stages {
        stage('Build') {
            steps {
                sh '''rm -f stldevs
rm -f run
go build cmd/stldevs/stldevs.go 
go build cmd/run/runmain.go 
./test.sh
mv runmain run'''
            }
        }
        stage('Deploy') {
            steps {
                sh '''
scp stldevs deploy@stldevs.com:~
scp run deploy@stldevs.com:~
ssh deploy@stldevs.com << EOF
   sudo service stldevs stop
   mv -f ~/stldevs /opt/stldevs
   mv -f ~/run /opt/stldevs
   cd /opt/stldevs
   chmod +x stldevs
   chmod +x run
   sudo service stldevs start
'''
            }
        }
    }
}
