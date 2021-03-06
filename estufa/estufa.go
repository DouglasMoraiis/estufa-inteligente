package main

import (
	"../camada"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/google/gopacket"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

type Sensor struct {
	nome string
	id uint16
	valor int16
}

type Atuador struct {
	nome string
	id uint16
	status uint16
}

type Estufa struct {
	sensores [] Sensor
	atuadores [] Atuador
}

// OS DADOS DOS Sensores e Atuadores SAO RECEBIDOS EM MEMORIA LOCAL PARA USO
var estufa Estufa
//-------------

func main()  {
	presetSensores()
	presetAtuadores()

	tcpAddrSensor, err := net.ResolveTCPAddr("tcp", ":3300")
	checkError(err, "ResolveTCPAddr")

	fmt.Println("-------- ESTUFA INICIADA ---------")
	fmt.Println("## Os Sensores iniciam com os valores: ")
	fmt.Println("## TEMP: 0, UMIDAD: 50, NIVELCO2: 50   ")
	fmt.Println("As leituras estão sendo enviadas...")

	go connRetornaSensoresInfo(tcpAddrSensor)

	for {
		continue
	}
}

func presetSensores() {
	//DEFININDO VALORES INICIAIS PARA A STRUCT SENSORES
	var temperatura Sensor
	temperatura.nome = "Temperatura"
	temperatura.id = 1
	temperatura.valor = 0

	var umidade Sensor
	umidade.nome = "Umidade do Solo"
	umidade.id = 2
	umidade.valor = 50

	var nivelCO2 Sensor
	nivelCO2.nome = "Nível de CO2"
	nivelCO2.id = 3
	nivelCO2.valor = 50

	//ADICIONANDO SENSORES A ESTUFA
	estufa.sensores = append(estufa.sensores, temperatura)
	estufa.sensores = append(estufa.sensores, umidade)
	estufa.sensores = append(estufa.sensores, nivelCO2)
	//----------------
}

func presetAtuadores() {
	//DEFININDO VALORES INICIAIS PARA A STRUCT ATUADORES
	var aquecedor Atuador
	aquecedor.nome = "Aquecedor"
	aquecedor.id = 10
	aquecedor.status = 0

	var resfriador Atuador
	resfriador.nome = "Resfriador"
	resfriador.id = 20
	resfriador.status = 0

	var irrigador Atuador
	irrigador.nome = "Irrigador"
	irrigador.id = 30
	irrigador.status = 0

	var injetorCO2 Atuador
	injetorCO2.nome = "InjetorCO2"
	injetorCO2.id = 40
	injetorCO2.status = 0

	//ADICIONANDO SENSORES A ESTUFA
	estufa.atuadores = append(estufa.atuadores, aquecedor)
	estufa.atuadores = append(estufa.atuadores, resfriador)
	estufa.atuadores = append(estufa.atuadores, irrigador)
	estufa.atuadores = append(estufa.atuadores, injetorCO2)
}

func connRetornaSensoresInfo(addr *net.TCPAddr) {
	for {
		conn, err := net.DialTCP("tcp", nil, addr)
		checkError(err, "DialTCP")

		var buffer bytes.Buffer

		buffer = converteSensoresEmArrayDeBytes(estufa.sensores, buffer)

		// O PACOTE É CRIADO DE ACORDO COM SensoresLayer EXIGE
		var pacote = gopacket.NewPacket(
			buffer.Bytes(),
			camada.SensoresLayerType,
			gopacket.Default,
		)

		// O PACOTE COM OS SENSORES ATUALIZADOS É ENVIADO PARA O SERVIDOR
		_,_ = conn.Write(pacote.Data())

		// O SERVIDOR MANDOU OS ATUADORES QUE FORAM ALTERADOS
		result, _ := ioutil.ReadAll(conn)

		if len(result) != 0  {

			tam, atuadores := conectaAtuadores(result[:])
			fmt.Println("------------ ATUADOR(ES) ALTERADO(S) --------------")
			for i := 0; i < tam; i++ {
				var status string
				if atuadores[i].status == 0 {
					status = "desligado."
				} else {
					status = "ligado."
				}
				s := strings.TrimSpace(atuadores[i].nome)
				fmt.Println("O " + s + " foi " + status)

				//ATUALIZA STATUS DOS ATUADORES NA MEMORIA LOCAL DA ESTUFA - NECESSARIO PARA O MUDANÇAS
				for j := 0; j < len(estufa.atuadores); j++ {
					if atuadores[i].id == estufa.atuadores[j].id {
						estufa.atuadores[j].status = atuadores[i].status
					}
				}
			}
		}

		// INCREMENTA AO VALOR DO PROXIMO SENSOR +3 OU -3 RANDOM, CASO ATUADOR NAO ESTIVER AGINDO
		simulaMudancas()
		_ = conn.Close()
		time.Sleep(1 * time.Second)
	}
}

func simulaMudancas() {
	if estufa.atuadores[0].status == 1 {
		estufa.sensores[0].valor++
	}else{
		if estufa.atuadores[1].status == 0 {
			unidade := randomMaisOuMenos()
			estufa.sensores[0].valor += unidade
		}
	}

	if estufa.atuadores[1].status == 1 {
		estufa.sensores[0].valor--
	}else{
		if estufa.atuadores[0].status == 0 {
			unidade := randomMaisOuMenos()
			estufa.sensores[0].valor += unidade
		}
	}

	if estufa.atuadores[2].status == 1 {
		estufa.sensores[1].valor++
	}else{
		unidade := randomMaisOuMenos()
		if (estufa.sensores[1].valor + unidade) >= 0 {
			estufa.sensores[1].valor += unidade
		}
	}

	if estufa.atuadores[3].status == 1 {
		estufa.sensores[2].valor++
	}else{
		unidade := randomMaisOuMenos()
		if (estufa.sensores[2].valor + unidade) >= 0 {
			estufa.sensores[2].valor += unidade
		}
	}
}

func randomMaisOuMenos() int16 {
	v := rand.Intn(5)
	if v % 2 == 0 {
		return 1*3
	}else{
		return -1*3
	}
}

func converteSensoresEmArrayDeBytes(sensores []Sensor, buffer bytes.Buffer) bytes.Buffer {
	for i := 0; i < 3; i++ {
		var nomeBytes = make([]byte, 15)
		for i, j := range []byte(sensores[i].nome) {
			nomeBytes[i] = byte(j)
		}
		var idSensorBytes = make([]byte, 2)
		var valorBytes = make([]byte, 4)

		binary.BigEndian.PutUint16(idSensorBytes, sensores[i].id)
		binary.BigEndian.PutUint32(valorBytes, uint32(sensores[i].valor))

		buffer.Write(nomeBytes)
		buffer.Write(idSensorBytes)
		buffer.Write(valorBytes)
	}

	return buffer
}

func conectaAtuadores(result []byte) (int, []Atuador){
	var tam = int(binary.BigEndian.Uint16(result[:2]))

	var atuadores [] Atuador
	if tam > 0 {
		var atuador1 Atuador
		atuadores = append(atuadores, atuador1)

		atuadores[0].nome = string(result[2:17])
		atuadores[0].id = binary.BigEndian.Uint16(result[17:19])
		atuadores[0].status = binary.BigEndian.Uint16(result[19:21])
	}

	if tam > 1 {
		var atuador2 Atuador
		atuadores = append(atuadores, atuador2)

		atuadores[1].nome = string(result[21:36])
		atuadores[1].id = binary.BigEndian.Uint16(result[36:38])
		atuadores[1].status = binary.BigEndian.Uint16(result[38:40])
	}

	if tam > 2 {
		var atuador3 Atuador
		atuadores = append(atuadores, atuador3)

		atuadores[2].nome = string(result[40:55])
		atuadores[2].id = binary.BigEndian.Uint16(result[55:57])
		atuadores[2].status = binary.BigEndian.Uint16(result[57:59])
	}

	if tam > 3 {
		var atuador4 Atuador
		atuadores = append(atuadores, atuador4)

		atuadores[3].nome = string(result[59:74])
		atuadores[3].id = binary.BigEndian.Uint16(result[74:76])
		atuadores[3].status = binary.BigEndian.Uint16(result[76:78])
	}

	return tam, atuadores
}

func checkError(err error, msg string){
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro em " + msg, err.Error())
		os.Exit(1)
	}
}
