import React from 'react';
import 'bootstrap/dist/css/bootstrap.css';
import {Button, Card, Row, Col} from 'react-bootstrap'

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8000';

const Order = ({orderData, setChangeWaiter,
deleteSingleOrder, setChangeOrder}) => {
    const imageUrl = orderData && orderData.image 
        ? (orderData.image.startsWith('http') ? orderData.image : `${API_BASE_URL}${orderData.image}`)
        : null;

    return(
        <Card className='customCard'>
            <Row className="align-items-center">
                <Col xs={2} className='customColumn text-center'>
                    {imageUrl ? (
                        <img 
                            src={imageUrl} 
                            alt={orderData.dish || 'Media'} 
                            style={{ width: '50px', height: '50px', objectFit: 'cover', borderRadius: '8px', border: '1px solid #EF6C00' }} 
                        />
                    ) : (
                        <span style={{ fontSize: '12px', color: '#888' }}>No Media</span>
                    )}
                </Col>
                <Col className='customColumn'>Dish: { orderData !== undefined && orderData.dish}</Col>
                <Col className='customColumn'>Server: { orderData !== undefined && orderData.server}</Col>
                <Col className='customColumn'>Table: { orderData !== undefined && orderData.table }</Col>
                <Col className='customColumn'>Price: ${ orderData !== undefined && orderData.price}</Col>
                <Col className='customColumn'><Button className='customDeleteButton' onClick={() => deleteSingleOrder(orderData._id)}>Delete Order</Button></Col>
                <Col className='customColumn'><Button className= 'customChangeWaiterButton' onClick={() => changeWaiter()}>Change Waiter</Button></Col>
                <Col className='customColumn'><Button className='customChangeOrderButton' onClick={() => changeOrder()}>Change Order</Button></Col>
            </Row>
        </Card>
    )

    function changeWaiter(){
        setChangeWaiter({
            "change":true,
            "id": orderData._id
        })
    }

    function changeOrder(){
        setChangeOrder({
            "change": true,
            "id": orderData._id
        })
    }

}

export default Order